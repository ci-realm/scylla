package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	macaron "gopkg.in/macaron.v1"

	"github.com/Jeffail/tunny"
)

func postHooksGithub(ctx *macaron.Context, Hook GithubHook) {
	if ctx.Req.Header.Get("X-Github-Event") == "pull_request" {
		go processGithub(pool, &Hook, progressHost(ctx))
	}

	ctx.JSON(200, map[string]string{"status": "OK"})
}

func progressHost(ctx *macaron.Context) string {
	proto := ctx.Req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	return fmt.Sprintf("%s://%s", proto, ctx.Req.Host)
}

func processGithub(pool *tunny.Pool, hook *GithubHook, host string) {
	j := &githubJob{Hook: hook, Host: host}
	j.status("pending", "Queueing...")
	_, err := pool.ProcessTimed(j, time.Minute*30)
	if err == tunny.ErrJobTimedOut {
		j.status("error", "Timeout after 30 minutes")
		log.Printf("Build of %s %s timed out\n", j.cloneURL(), j.sha())
	}
}

type githubJob struct {
	Hook *GithubHook
	Host string
}

func newGithubJobFromJSONFile(path string) *githubJob {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Couldn't open file %s: %s\n", path, err)
		return nil
	}

	job := &githubJob{Hook: &GithubHook{}}
	if err = json.NewDecoder(file).Decode(job.Hook); err != nil {
		log.Printf("Failed to decode JSON %s: %s\n", path, err)
		return nil
	}
	return job
}

var sanitizeUrlPath = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

func (j *githubJob) saneFullName() string {
	return sanitizeUrlPath.ReplaceAllString(j.Hook.Repository.FullName, "_")
}

func (j *githubJob) targetURL() string {
	uri, _ := url.Parse(j.Host)
	uri.Path = fmt.Sprintf("/builds/%s/%s", j.saneFullName(), j.sha())
	return uri.String()
}

func (j *githubJob) cloneURL() string {
	return j.Hook.PullRequest.Head.Repo.CloneURL
}

func (j *githubJob) sha() string {
	return j.Hook.PullRequest.Head.Sha
}

func (j *githubJob) pname() string {
	return j.saneFullName() + "-" + j.sha()
}

func (j *githubJob) rootDir() string {
	return "./ci"
}

func (j *githubJob) buildDir() string {
	return cleanJoin(j.rootDir(), j.saneFullName(), j.sha())
}

func (j *githubJob) sourceDir() string {
	return filepath.Join(j.buildDir(), "source")
}

func (j *githubJob) resultLink() string {
	return filepath.Join(j.buildDir(), "result")
}

func (j *githubJob) ciNixPath() string {
	return filepath.Join(j.buildDir(), "source", "ci.nix")
}

func (j *githubJob) clone() {
	j.status("pending", "Cloning...")

	runCmd(exec.Command(
		"git", "clone", j.cloneURL(), j.sourceDir()))

	j.status("pending", "Checkout...")

	runCmd(exec.Command(
		"git",
		"-c", "advice.detachedHead=false",
		"-C", j.sourceDir(),
		"checkout", j.sha()))
}

func (j *githubJob) nix(subcmd string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	return runCmd(exec.Command(
		"nix",
		append([]string{
			subcmd,
			"--allow-import-from-derivation",
			"--auto-optimise-store",
			"--enforce-determinism",
			"--fallback",
			"--http2",
			"--keep-build-log",
			"--restrict-eval",
			"--show-trace",
			"--max-build-log-size", "10000000",
			"--max-silent-time", "30",
			"--timeout", "30",
			"--option", "allowed-uris", "https://github.com/ https://source.xing.com/",
			"-I", "./nix",
			"-I", j.sourceDir(),
			"--argstr", "pname", j.pname(),
		}, args...)...,
	))
}

func (j *githubJob) nixLog() {
	j.nix("log", "-f", j.ciNixPath(), "")
}

func (j *githubJob) nixBuild() {
	j.status("pending", "Nix Build...")
	stdout, stderr, err := j.nix(
		"build", "--out-link", j.resultLink(), "-f", j.ciNixPath())

	j.writeOutput(stdout, stderr)

	if err != nil {
		j.status("failure", err.Error())
	} else {
		j.status("success", "Evaluation succeeded")
	}
}

func (j *githubJob) writeOutput(stdout, stderr *bytes.Buffer) {
	j.writeOutputToFile("stdout", stdout)
	j.writeOutputToFile("stderr", stderr)
}

func (j *githubJob) writeOutputToFile(baseName string, output *bytes.Buffer) {
	pathName := filepath.Join(j.buildDir(), baseName)
	file, err := os.Create(pathName)
	if err != nil {
		log.Printf("Failed to create file %s: %s\n", pathName, err)
		return
	}
	defer file.Close()
	_, err = output.WriteTo(file)
	if err != nil {
		log.Printf("Failed to write file %s: %s\n", pathName, err)
	}
}

func (j *githubJob) persistHook() {
	pathName := filepath.Join(j.buildDir(), "hook.json")
	file, err := os.Create(pathName)
	if err != nil {
		log.Printf("Failed to create file %s: %s\n", pathName, err)
		return
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(j.Hook)
	if err != nil {
		log.Printf("Failed to write file %s: %s\n", pathName, err)
	}
}

func (j *githubJob) build() string {
	log.Printf("Starting work on %s %s...", j.cloneURL(), j.sha())

	_ = os.RemoveAll(j.sourceDir())

	j.clone()
	j.persistHook()
	j.nixBuild()

	_ = os.RemoveAll(j.sourceDir())

	return "OK"
}

func (j *githubJob) status(state, description string) {
	setGithubStatus(
		j.targetURL(),
		j.Hook.PullRequest.StatusesURL,
		state,
		description,
	)
}

func setGithubStatus(targetURL, statusURL, state, description string) {
	if len(description) > 138 {
		description = description[0:138]
	}

	status := map[string]string{
		"state":       state,
		"target_url":  targetURL,
		"description": description,
		"context":     "Scylla",
	}
	body := &bytes.Buffer{}

	json.NewEncoder(body).Encode(&status)

	req, err := http.NewRequest("POST", statusURL, body)
	if err != nil {
		log.Fatalf("Failed creating request: %s", err)
	}

	req.SetBasicAuth(config.GithubUser, config.GithubToken)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error while calling Github API: %s", err)
	}
}

func cleanJoin(parts ...string) string {
	return filepath.Clean(filepath.Join(parts...))
}
