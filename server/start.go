package server

import (
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	macaron "gopkg.in/macaron.v1"
)

var templateFuncMap = []template.FuncMap{{
	"FormatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"ToClass": func(s string) string {
		return strings.ToLower(s)
	},
	"ShortSHA": func(s string) string {
		return s[0:7]
	},
	"FormatDuration": func(s time.Duration) string {
		return s.String()
	},
	"FormatTimeAgo": func(s time.Time) string {
		return time.Since(s).String()
	},
	"ScyllaVersionLink": func() string {
		return "https://source.xing.com/e-recruiting-api-team/scylla"
	},
	"ScyllaHostname": func() string {
		if host := os.Getenv("HOSTNAME"); host != "" {
			return host
		}
		return "localhost"
	},
}}

func Start() {
	ParseConfig()
	SetupDB()
	populateKnownHosts()
	defer pgxpool.Close()

	go SetupQueue()

	go startLogDistributor(pgxpool)

	m := macaron.Classic()
	m.SetAutoHead(true)
	m.NotFound(func(ctx *macaron.Context) {
		ctx.HTML(404, "not_found")
	})
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Layout:     "layout",
		Extensions: []string{".html"},
		Funcs:      templateFuncMap,
	}))

	setupRouting(m)
	m.Run(config.Host, config.Port)
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
		cmd := exec.Command("ssh-keyscan", "-p", "443", host)
		output, err := cmd.Output()
		if err != nil {
			logger.Fatalln("Couldn't get host key", err)
		}
		hostKey := strings.TrimSpace(string(output))
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
