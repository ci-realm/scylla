package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
)

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

func ParseConfig() {
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
