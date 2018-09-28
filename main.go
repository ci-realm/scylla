package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	macaron "gopkg.in/macaron.v1"

	arg "github.com/alexflint/go-arg"
	que "github.com/bgentry/que-go"
	"github.com/jackc/pgx"
)

func init() {
	err := os.MkdirAll(os.TempDir(), os.FileMode(0755))
	if err != nil {
		logger.Fatalln("failed making ", os.TempDir(), err)
	}
}

var logger = log.New(os.Stderr, "[scylla] ", log.Lshortfile|log.Ltime|log.Ldate|log.LUTC)

var config struct {
	BuildDir string `arg:"--build-dir,env:BUILD_DIR"`
	Builders string `arg:"--builders,required,env:BUILDERS"`
	// BuildersPrivateKey string `arg:"--builders-private-key,env:BUILDERS_PRIVATE_KEY"`
	GithubToken       string `arg:"--github-token,required,env:GITHUB_TOKEN"`
	GithubUrl         string `arg:"--github-url,required,env:GITHUB_URL"`
	GithubUser        string `arg:"--github-user,required,env:GITHUB_USER"`
	Host              string `arg:"--host,env:HOST"`
	Port              int    `arg:"--port,env:PORT"`
	PrepareKnownHosts bool   `arg:"--prepare-known-hosts,env:PREPARE_KNOWN_HOSTS"`
	PrivateSSHKey     string `arg:"--private-ssh-key,required,env:PRIVATE_SSH_KEY"`
	DatabaseURL       string `arg:"--database-url,required,env:DATABASE_URL"`
}

func main() {
	parseConfig()
	populateKnownHosts()
	setupQueue()

	defer pgxpool.Close()

	m := macaron.Classic()
	m.SetAutoHead(true)
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Layout:     "layout",
		Extensions: []string{".html"},
		Funcs: []template.FuncMap{
			{"FormatTime": func(t time.Time) string {
				return t.Format(time.RFC1123)
			}},
		},
	}))

	setupRouting(m)
	m.Run(config.Host, config.Port)

	// mux := http.NewServeMux()
	// mux.Handle("/", m)

	// graceful.Run(config.Host+":"+config.Port, 2*time.Second, mux)
}

var (
	queueClient *que.Client
	pgxpool     *pgx.ConnPool
)

func setupQueue() {
	pgxcfg, err := pgx.ParseURI(config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	pgxpool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:   pgxcfg,
		AfterConnect: que.PrepareStatements,
	})
	if err != nil {
		log.Fatal(err)
	}

	queueClient = que.NewClient(pgxpool)
	workMap := que.WorkMap{"GithubPR": runGithubPR}
	workers := que.NewWorkerPool(queueClient, workMap, runtime.NumCPU())
	workers.Start()
}

// ssh-keyscan <host> >> ~/.ssh/known_hosts
func populateKnownHosts() {
	if !config.PrepareKnownHosts {
		return
	}

	knownHosts := []string{}
	hosts := []string{}

	// TODO: new builders syntax is:
	// --builders 'ssh://mac x86_64-darwin ; ssh://beastie x86_64-freebsd'
	for _, line := range strings.Split(config.Builders, ";") {
		words := strings.Split(line, " ")
		if len(words) < 1 {
			logger.Fatalln("At least one builder must be specified")
		}
		userAndHost := words[0]
		words = strings.Split(userAndHost, "@")
		if len(words) < 1 {
			logger.Fatalln("At least one builder must be specified")
		}
		host := words[len(words)-1]
		hosts = append(hosts, host)
		stdout, _, err := runCmd(exec.Command("ssh-keyscan", "-p", "443", host))
		if err != nil {
			logger.Fatalln("Couldn't get host key", err)
		}
		hostKey := strings.TrimSpace(stdout.String())
		logger.Println("Adding to known_hosts:", hostKey)
		knownHosts = append(knownHosts, hostKey)
	}

	writeFile("/.ssh/config", sshConfigFor(hosts), 0600)
	writeFile("/.ssh/known_hosts", strings.Join(knownHosts, "\n"), 0600)
	writeFile("/id_ed25519", config.PrivateSSHKey+"\n", 0600)

	// FIXME: hack until the image builder is fixed again
	writeFile("/etc/passwd", "root:x:0:0:root:/:/bin/sh", 0644)
	writeFile("/etc/nix/nix.conf", "build-users-group =", 0600)
}

func sshConfigFor(hosts []string) string {
	content := []string{}
	for _, host := range hosts {
		content = append(content, "Host "+host, "Port 443", "User root")
	}
	return strings.Join(content, "\n")
}

func writeFile(dest, content string, mode os.FileMode) {
	logger.Println("Writing", dest)
	dir := filepath.Dir(dest)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		logger.Fatalln("Couldn't create directory:", dir, err)
	}

	err = ioutil.WriteFile(dest, []byte(content), mode)
	if err != nil {
		logger.Fatalln("Couldn't write file:", dest, err)
	}
}

func parseConfig() {
	config.Host = "0.0.0.0"
	config.Port = 8080
	config.BuildDir = "./ci"

	arg.MustParse(&config)

	if strings.HasPrefix(config.GithubUser, "/") {
		if content, err := ioutil.ReadFile(config.GithubUser); err != nil {
			config.GithubUser = string(content)
		}
	}

	if strings.HasPrefix(config.GithubToken, "/") {
		if content, err := ioutil.ReadFile(config.GithubToken); err != nil {
			config.GithubToken = string(content)
		}
	}
}

func worker(work interface{}) interface{} {
	switch w := work.(type) {
	case *githubJob:
		return w.build()
	}

	return "Couldn't find work type"
}

// runCmd returns stdout, stderr, and any errors
func runCmd(cmd *exec.Cmd) (*bytes.Buffer, *bytes.Buffer, error) {
	logger.Printf("%s %v\n", cmd.Path, cmd.Args)

	var stdoutBuf, stderrBuf bytes.Buffer

	cmd.Stdout = io.MultiWriter(&stdoutBuf, os.Stdout)
	cmd.Stderr = io.MultiWriter(&stderrBuf, os.Stderr)

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	if err := cmd.Wait(); err != nil {
		return &stdoutBuf, &stderrBuf, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	return &stdoutBuf, &stderrBuf, nil
}
