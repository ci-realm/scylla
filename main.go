package main

import (
	"bufio"
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
	"sync"
	"time"

	macaron "gopkg.in/macaron.v1"

	arg "github.com/alexflint/go-arg"
	"github.com/jackc/pgx"
	"github.com/manveru/scylla/queue"
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
	PrivateSSHKeyPath string `arg:"--private-ssh-key-path,env:PRIVATE_SSH_KEY_PATH"`
	DatabaseURL       string `arg:"--database-url,required,env:DATABASE_URL"`
}

func main() {
	parseConfig()
	populateKnownHosts()
	setupDB()
	go setupQueue()

	defer pgxpool.Close()

	m := macaron.Classic()
	m.SetAutoHead(true)
	m.NotFound(func(ctx *macaron.Context) {
		ctx.HTML(404, "not_found")
	})
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Layout:     "layout",
		Extensions: []string{".html"},
		Funcs: []template.FuncMap{{
			"FormatTime": func(t time.Time) string {
				return t.Format("2006-01-02 15:04:05")
			},
			"ToClass": func(s string) string {
				return strings.ToLower(s)
			},
			"ShortSHA": func(s string) string { return s },
			"FormatDuration": func(s time.Duration) string {
				return s.String()
			},
			"FormatTimeAgo": func(s time.Time) string {
				return time.Now().Sub(s).String()
			},
		}},
	}))

	setupRouting(m)
	m.Run(config.Host, config.Port)
}

func parseConfig() {
	config.Host = "0.0.0.0"
	config.Port = 8080
	config.BuildDir = "./ci"
	config.PrivateSSHKeyPath = "/id_ed25519"

	err := arg.Parse(&config)
	if err != nil { // needed for goconvey
		if strings.HasPrefix(err.Error(), "unknown argument -test.v") {
			return
		}
		if strings.HasPrefix(err.Error(), "unknown argument -test.coverprofile") {
			return
		}
		fmt.Println(err)
		os.Exit(1)
	}

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

func setupDB() {
	pgxcfg, err := pgx.ParseURI(config.DatabaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	if strings.Contains(config.DatabaseURL, "amazonaws.com") {
		tunnelStarted := make(chan bool)
		go setupDatabaseTunnel(tunnelStarted)
		<-tunnelStarted
		pgxcfg.Port = localDBPort
		pgxcfg.Host = "127.0.0.1"
	}

	pgxpool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     pgxcfg,
		AfterConnect:   func(*pgx.Conn) error { return nil },
		MaxConnections: 50,
	})
	if err != nil {
		logger.Fatalln(err)
	}
}

var (
	jobQueue *queue.Queue
	pgxpool  *pgx.ConnPool
)

func setupQueue() {
	logger.Println("Setting up worker queue")

	jobQueue = &queue.Queue{
		Timeout:    time.Hour,
		CheckEvery: time.Second * 10,
		Pool:       pgxpool,
		Name:       "scylla",
		Retries:    3,
	}

	err := jobQueue.Start(runtime.NumCPU(), func(item *queue.Item) error {
		return runGithubPR(item)
	})
	if err != nil {
		logger.Fatalln(err)
	}
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
		output, err := runCmd(exec.Command("ssh-keyscan", "-p", "443", host))
		if err != nil {
			logger.Fatalln("Couldn't get host key", err)
		}
		hostKey := strings.TrimSpace(output.String())
		logger.Println("Adding to known_hosts:", hostKey)
		knownHosts = append(knownHosts, hostKey)
	}

	writeFile("/.ssh/config", sshConfigFor(hosts), 0600)
	writeFile("/.ssh/known_hosts", strings.Join(knownHosts, "\n"), 0600)
	writeFile(config.PrivateSSHKeyPath, config.PrivateSSHKey+"\n", 0600)

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

// runCmd returns stdout, stderr, and any errors
func runCmd(cmd *exec.Cmd) (*bytes.Buffer, error) {
	logger.Printf("%s %v\n", cmd.Path, cmd.Args)

	var combinedOutput bytes.Buffer

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Fatalln(err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Fatalln(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	onLine := make(chan string, 100)
	go func(output *bytes.Buffer) {
		lineLogger := log.New(output, "["+filepath.Base(cmd.Path)+"] ", log.Ldate|log.Ltime|log.LUTC)
		for line := range onLine {
			logger.Println(line)
			lineLogger.Println(line)
		}
	}(&combinedOutput)
	go logPipe(wg, stderrPipe, onLine)
	go logPipe(wg, stdoutPipe, onLine)

	if err := cmd.Start(); err != nil {
		io.WriteString(&combinedOutput, err.Error())
		return &combinedOutput, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	wg.Wait()
	close(onLine)

	if err := cmd.Wait(); err != nil {
		io.WriteString(&combinedOutput, err.Error())
		return &combinedOutput, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	return &combinedOutput, nil
}

func logPipe(wg *sync.WaitGroup, input io.ReadCloser, onLine chan string) {
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		onLine <- scanner.Text()
	}
	wg.Done()
}

func withConn(ctx *macaron.Context, f func(*pgx.Conn) error) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		ctx.Error(500, err.Error())
		return
	}
	defer pgxpool.Release(conn)
	err = f(conn)
	if err != nil {
		ctx.Error(500, err.Error())
	}
}
