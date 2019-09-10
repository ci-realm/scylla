package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
)

var config struct {
	BuildDir    string `arg:"--build-dir,env:BUILD_DIR" help:"location for git checkouts"`
	Builders    string `arg:"--builders,env:BUILDERS" help:"nix.conf syntax"`
	CachixName  string `arg:"--cachix-name,env:CACHIX_NAME" help:"Set to push results to cachix"`
	DatabaseURL string `arg:"--database-url,required,env:DATABASE_URL" help:"postgresql://user:pass@host:port/db"`
	GithubToken string `arg:"--github-token,required,env:GITHUB_TOKEN" help:"Token for GitHub auth"`
	GithubUrl   string `arg:"--github-url,required,env:GITHUB_URL" help:"base url for GitHub"`
	GithubUser  string `arg:"--github-user,required,env:GITHUB_USER" help:"User for GitHub auth"`
	Host        string `arg:"--host,env:HOST" help:"Host for listening"`
	NixCopyURL  string `arg:"--nix-copy-url,env:NIX_COPY_URL" help:"Set to nix copy results"`
	Port        int    `arg:"--port,env:PORT" help:"Listen on port"`
	Mode        string `arg:"--mode,env:SCYLLA_MODE" help:"set to development to proxy to frontend server"`
}

func ParseConfig() {
	config.Host = "0.0.0.0"
	config.Port = 8080
	config.BuildDir = "./ci"
	config.GithubUrl = "https://github.com"

	parser, err := arg.NewParser(arg.Config{Program: "scylla"}, &config)
	if err != nil {
		logger.Fatal(err)
	}

	err = parser.Parse(os.Args[1:])
	if err != nil { // needed for goconvey
		if strings.HasPrefix(err.Error(), "unknown argument -test.") {
			return
		}

		if err == arg.ErrHelp {
			parser.WriteHelp(os.Stdout)
			os.Exit(0)
		}

		if err == arg.ErrVersion {
			fmt.Println("scylla version 0.0.1")
			os.Exit(0)
		}

		parser.WriteUsage(os.Stdout)
		fmt.Println(err)
		os.Exit(1)
	}

	if strings.HasPrefix(config.GithubUser, "/") {
		logger.Printf("Loading github user from file %s\n", config.GithubUser)
		content, err := ioutil.ReadFile(config.GithubUser)
		if err != nil {
			logger.Fatal(err)
		}

		config.GithubUser = string(content)
	}

	if strings.HasPrefix(config.GithubToken, "/") {
		logger.Printf("Loading github token from file %s\n", config.GithubToken)
		content, err := ioutil.ReadFile(config.GithubToken)
		if err != nil {
			logger.Fatal(err)
		}
		config.GithubToken = string(content)
	}
}
