package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	macaron "gopkg.in/macaron.v1"

	que "github.com/bgentry/que-go"
	"github.com/jackc/pgx"
)

type githubJob struct {
	Hook    *GithubHook
	Host    string
	buildID int
	conn    *pgx.Conn
}

type githubJobItem struct {
	BuildID int
	Host    string
}

func postHooksGithub(ctx *macaron.Context, hook GithubHook) {
	if ctx.Req.Header.Get("X-Github-Event") == "pull_request" && hook.Action != "closed" {
		if err := enqueueGithub(&hook, progressHost(ctx)); err != nil {
			ctx.JSON(500, map[string]string{"status": "ERROR", "error": err.Error()})
			return
		}

		ctx.JSON(200, map[string]string{"status": "OK"})
		return
	}

	ctx.JSON(200, map[string]string{"status": "IGNORED"})
}

func enqueueGithub(hook *GithubHook, host string) error {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Fatalln(err)
	}
	defer pgxpool.Release(conn)

	ghJob := githubJob{Host: host, Hook: hook, conn: conn}

	projectID, err := ghJob.findOrCreateProjectID()
	logger.Println("projectID:", projectID, err)
	if err != nil {
		ghJob.onError(err, "project not available")
		return err
	}

	buildID, err := ghJob.createBuild(projectID)
	if err != nil {
		ghJob.onError(err, "couldn't create build")
		return err
	}

	ghJobItem := &githubJobItem{BuildID: buildID, Host: host}
	args, err := json.Marshal(ghJobItem)
	if err != nil {
		return err
	}

	err = queueClient.Enqueue(&que.Job{Type: "GithubPR", Args: args})
	if err != nil {
		logger.Println("Enqueue:", err)
	}
	return err
}

func runGithubPR(j *que.Job) error {
	jobItem := &githubJobItem{}
	err := json.Unmarshal(j.Args, &jobItem)
	if err != nil {
		logger.Println(err)
		return err
	}

	conn := j.Conn()

	job, err := findBuildByID(conn, jobItem.BuildID)
	if err != nil {
		logger.Println(err)
		return err
	}

	job.conn = conn
	job.Host = jobItem.Host

	err = job.build()
	logger.Println(err)
	return err
}

func (j *githubJob) build() error {
	lockID, err := strconv.ParseInt(j.sha()[0:16], 16, 64)
	if err != nil {
		return err
	}

	logger.Printf("%s: Waiting for lock %d...\n", j.id(), lockID)

	ctx, _ := context.WithTimeout(context.Background(), time.Minute*10)
	txn, err := pgxpool.BeginEx(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable})
	defer txn.Rollback()

	_, err = txn.Exec(`SET LOCAL lock_timeout = '60s';`)
	if err != nil {
		return err
	}

	_, err = txn.Exec(`SELECT pg_advisory_xact_lock($1);`, lockID)
	if err != nil {
		return err
	}

	if len(j.resultNixPaths()) > 0 {
		logger.Printf("%s: skipping build, results exist already.\n", j.id())
		return nil
	}

	logger.Printf("%s: Starting work...\n", j.id())

	fd, err := os.Open(j.sourceDir())
	_ = fd.Close()
	if err == nil {
		if err := j.gitFetch(); err != nil {
			j.onError(err, "failed fetching "+j.cloneURL())
			return err
		}
	} else {
		if err := j.gitClone(); err != nil {
			j.onError(err, "failed cloning "+j.cloneURL())
			return err
		}
	}

	txn.Rollback()

	return j.nixBuild()
}

func (j *githubJob) onQueue() error {
	updateBuildStatus(j, "queue")
	j.status("pending", "Queued")
	logger.Printf("Queued build of %s\n", j.id())

	return nil
}

func (j *githubJob) onTimeout(timeout time.Duration) {
	j.status("error", "Timeout after "+timeout.String()+" minutes")
	logger.Printf("Build of %s timed out\n", j.id())
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
	return config.BuildDir
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

func (j *githubJob) gitFetch() error {
	j.status("pending", "Fetching...")

	githubAuth := githubAuthKey(config.GithubUrl, config.GithubToken) + "=" + config.GithubUrl

	_, _, err := runCmd(exec.Command(
		"git", "-c", githubAuth, "-C", j.sourceDir(), "fetch"))

	return err
}

func (j *githubJob) gitClone() error {
	j.status("pending", "Cloning...")

	githubAuth := githubAuthKey(config.GithubUrl, config.GithubToken) + "=" + config.GithubUrl

	_, _, err := runCmd(exec.Command(
		"git", "clone", "-c", githubAuth, j.cloneURL(), j.sourceDir()))

	if err != nil {
		return err
	}

	j.status("pending", "Checkout...")

	logger.Println("before exec")

	_, _, err = runCmd(exec.Command(
		"git", "-c", "advice.detachedHead=false", "-C", j.sourceDir(), "checkout", j.sha()))

	logger.Println("git checkout result:", err)

	return err
}

func (j *githubJob) nix(subcmd string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	return runCmd(exec.Command(
		"nix",
		append([]string{
			subcmd,
			"--show-trace",
			"--builders", config.Builders,
			"--max-jobs", "0", // force remote builds
			"-I", "./nix",
			"-I", j.sourceDir(),
			"--argstr", "pname", j.pname(),
		}, args...)...,
	))
}

func (j *githubJob) nixLog() (string, string, error) {
	sout, serr, err := j.nix("log", "-f", j.ciNixPath(), "")
	if err == nil {
		return sout.String(), serr.String(), err
	}

	stderrPath := filepath.Join(j.buildDir(), "stderr")
	stderrBytes, err := ioutil.ReadFile(stderrPath)
	if err != nil {
		return "", "", errors.New("No trace of logs found")
	}

	drvs := parseDrvsFromStderr(stderrBytes)
	for _, drv := range drvs {
		sout, serr, err = runCmd(exec.Command("nix", "log", drv))
	}

	return sout.String(), serr.String(), err
}

var matchFailine = regexp.MustCompile(`error: build of .+ failed`)
var matchFailDrvs = regexp.MustCompile(`[^'\s]+\.drv`)

func parseDrvsFromStderr(input []byte) []string {
	line := matchFailine.FindString(string(input))
	return matchFailDrvs.FindAllString(line, -1)
}

func (j *githubJob) nixBuild() error {
	updateBuildStatus(j, "build")
	j.status("pending", "Nix Build...")

	stdout, stderr, err := j.nix(
		"build", "--out-link", j.resultLink(), "-f", j.ciNixPath())

	j.writeOutput(stdout, stderr)

	if err != nil {
		j.status("failure", err.Error())
		updateBuildStatus(j, "failure")
		return errors.New("Nix Failure: " + err.Error())
	}

	j.onSuccess()

	return nil
}

func (j *githubJob) writeOutput(stdout, stderr *bytes.Buffer) {
	j.writeOutputToFile("stdout", stdout)
	j.writeOutputToFile("stderr", stderr)
	j.writeOutputToDB("stdout", stdout)
	j.writeOutputToDB("stderr", stderr)
}

func (j *githubJob) writeOutputToFile(baseName string, output *bytes.Buffer) {
	pathName := filepath.Join(j.buildDir(), baseName)
	file, err := os.Create(pathName)
	if err != nil {
		logger.Printf("Failed to create file %s: %s\n", pathName, err)
		return
	}
	defer file.Close()
	_, err = output.WriteTo(file)
	if err != nil {
		logger.Printf("Failed to write file %s: %s\n", pathName, err)
	}
}

func (j *githubJob) writeOutputToDB(basename string, output *bytes.Buffer) {
	insertLog(j.buildID, basename, output.String())
}

func (j *githubJob) status(state, description string) {
	logger.Println(j.id(), ":", state, description)
	setGithubStatus(
		j.targetURL(),
		j.Hook.PullRequest.StatusesURL,
		state,
		description,
	)
}

func (j *githubJob) onError(err error, msg string) {
	logger.Printf("%s: %s: %s\n", j.id(), msg, err)
	j.status("error", fmt.Sprintf("%s: %s", msg, err))
	updateBuildStatus(j, "failure")

	_ = os.RemoveAll(j.sourceDir())
}

func (j *githubJob) onSuccess() {
	logger.Printf("%s: success\n", j.id())
	j.status("success", "Evaluation of "+j.id()+" succeeded")
	updateBuildStatus(j, "success")

	// TODO: also remove outputs to allow GC
	_ = os.RemoveAll(j.sourceDir())
	j.copyResultsToCache()
}

func (j *githubJob) resultNixPaths() []string {
	matches, err := filepath.Glob(cleanJoin(j.buildDir(), "result*"))
	if err != nil {
		j.onError(err, "failed enumerating results")
		return nil
	}

	for n, match := range matches {
		nixStorePath, err := filepath.EvalSymlinks(match)
		if err != nil {
			j.onError(err, "failed resolving result symlink")
			return nil
		}
		logger.Println("result:", nixStorePath)
		matches[n] = nixStorePath
	}

	return matches
}

func (j *githubJob) copyResultsToCache() {
	for _, nixStorePath := range j.resultNixPaths() {
		err := insertResult(j.buildID, nixStorePath)
		if err != nil {
			j.onError(err, "failed storing result in DB")
			return
		}
	}

	runCmd(exec.Command(
		"ssh", "root@3.120.166.103",
		"nix", "copy",
		"--all",
		"--to", "s3://scylla-cache?region=eu-central-1",
	))
}

func (j *githubJob) id() string {
	return j.cloneURL() + "/" + j.sha()
}

func (j *githubJob) findOrCreateProjectID() (int, error) {
	return findOrCreateProjectID(j.Hook.Repository.FullName)
}

func (j *githubJob) createBuild(projectID int) (int, error) {
	return insertBuild(j.conn, projectID, j)
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
		logger.Printf("Error while calling Github API: %s\n", err)
	}
}

func cleanJoin(parts ...string) string {
	return filepath.Clean(filepath.Join(parts...))
}

func progressHost(ctx *macaron.Context) string {
	proto := ctx.Req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	return fmt.Sprintf("%s://%s", proto, ctx.Req.Host)
}

func newGithubJobFromJSONFile(path string) *githubJob {
	file, err := os.Open(path)
	if err != nil {
		logger.Printf("Couldn't open file %s: %s\n", path, err)
		return nil
	}

	job := &githubJob{Hook: &GithubHook{}}
	if err = json.NewDecoder(file).Decode(job.Hook); err != nil {
		logger.Printf("Failed to decode JSON %s: %s\n", path, err)
		return nil
	}
	return job
}

func githubAuthKey(givenURL, token string) string {
	u, err := url.Parse(givenURL)
	if err != nil {
		logger.Fatalln("Couldn't parse github url", err)
	}
	u.User = url.UserPassword(token, "x-oauth-basic")
	return "url." + u.String() + ".insteadOf"
}
