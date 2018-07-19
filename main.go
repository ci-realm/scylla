package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	macaron "gopkg.in/macaron.v1"

	"github.com/Jeffail/tunny"
)

var serverURL = flag.String("url", "https://scylla.ngrok.io", "set host used for GH status links")

var pool *tunny.Pool

func main() {
	pool = tunny.NewFunc(runtime.NumCPU(), worker)

	defer pool.Close()

	m := macaron.Classic()
	m.SetAutoHead(true)
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Layout:     "layout",
		Extensions: []string{".html"},
	}))

	setupRouting(m)

	m.Run(8080)
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
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	var errStdout, errStderr error

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	if err := cmd.Wait(); err != nil {
		return &stdoutBuf, &stderrBuf, fmt.Errorf("%s failed with %s\n", cmd.Path, err)
	}

	if errStdout != nil || errStderr != nil {
		return &stdoutBuf, &stderrBuf, fmt.Errorf("failed to capture stdout or stderr\n")
	}

	return &stdoutBuf, &stderrBuf, nil
}