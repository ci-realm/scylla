package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	macaron "gopkg.in/macaron.v1"

	"github.com/Jeffail/tunny"
	arg "github.com/alexflint/go-arg"
)

var pool *tunny.Pool

var config struct {
	GithubUser  string `arg:"--github-user,required,env:GITHUB_USER"`
	GithubToken string `arg:"--github-token,required,env:GITHUB_TOKEN"`
	Host        string `arg:"env:HOST"`
	Port        int    `arg:"env:PORT"`
}

func main() {
	parseConfig()

	pool = tunny.NewFunc(runtime.NumCPU(), worker)

	defer pool.Close()

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

func parseConfig() {
	config.Host = "0.0.0.0"
	config.Port = 8080

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

func runCmd(cmd *exec.Cmd) (*bytes.Buffer, *bytes.Buffer, error) {
	log.Printf("%s %v", cmd.Path, cmd.Args)

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var stdoutBuf, stderrBuf bytes.Buffer

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	var errStdout, errStderr error

	go func() {
		_, errStdout = io.Copy(&stdoutBuf, stdoutIn)
	}()

	go func() {
		_, errStderr = io.Copy(&stderrBuf, stderrIn)
	}()

	if err := cmd.Wait(); err != nil {
		return &stdoutBuf, &stderrBuf, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	if errStdout != nil || errStderr != nil {
		return &stdoutBuf, &stderrBuf, fmt.Errorf("failed to capture stdout or stderr\n")
	}

	return &stdoutBuf, &stderrBuf, nil
}
