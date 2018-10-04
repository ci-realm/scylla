package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	macaron "gopkg.in/macaron.v1"

	"github.com/jackc/pgx"
	"github.com/manveru/scylla/queue"
)

type githubJob struct {
	Hook    *GithubHook
	Host    string
	buildID int
	conn    *pgx.Conn
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

	item := &queue.Item{Args: map[string]interface{}{"build_id": buildID, "Host": host}}
	return jobQueue.Insert(item)
}

func runGithubPR(j *queue.Item) error {
	if len(j.Errors) > 3 {
		logger.Printf("giving up on job %d after 3 tries", j.ID)
		return nil
	}

	args := j.Args.(map[string]interface{})
	host := args["Host"].(string)
	buildID := int(args["build_id"].(float64))

	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Println(err)
		return err
	}
	defer pgxpool.Release(conn)

	job, err := findBuildByID(conn, buildID)
	if err != nil {
		logger.Println(err)
		return err
	}

	job.conn = conn
	job.Host = host

	err = job.build()
	if err != nil {
		logger.Println(err)
	}
	return err
}

var crcTable *crc64.Table

func init() {
	crcTable = crc64.MakeTable(crc64.ECMA)
}

// lockID uses CRC, which should be enough for our short-lived DB locks
func (j *githubJob) lockID() int64 {
	ui64 := crc64.Checksum([]byte(j.sha()), crcTable)
	return int64(ui64)
}

func (j *githubJob) build() error {
	lockID := j.lockID()
	logger.Printf("%s: Waiting for lock %d...\n", j.id(), lockID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	txn, err := pgxpool.BeginEx(ctx, nil)
	if err != nil {
		cancel()
		return err
	}
	defer func() { cancel(); _ = txn.Rollback() }()

	_, err = txn.Exec(`SET LOCAL lock_timeout = '60s';`)
	if err != nil {
		return err
	}

	// TODO: ideally we want a machine-level lock instead.
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

	// once we get here, others can use this hash as well, as the source should be
	// immutable anyway.
	_ = txn.Rollback()

	if drvs, err := j.nixInstantiate(); err != nil {
		return err
	} else {
		logger.Println("drvs:", drvs)
	}

	return j.nixBuild()
}

func (j *githubJob) gitFetch() error {
	j.status("pending", "Fetching...")

	githubAuth := githubAuthKey(config.GithubUrl, config.GithubToken) + "=" + config.GithubUrl

	_, err := runCmd(exec.Command(
		"git", "-c", githubAuth, "-C", j.sourceDir(), "fetch"))

	return err
}

func (j *githubJob) gitClone() error {
	j.status("pending", "Cloning...")

	githubAuth := githubAuthKey(config.GithubUrl, config.GithubToken) + "=" + config.GithubUrl

	_, err := runCmd(exec.Command(
		"git", "clone", "-c", githubAuth, j.cloneURL(), j.sourceDir()))

	if err != nil {
		return err
	}

	j.status("pending", "Checkout...")

	_, err = runCmd(exec.Command(
		"git", "-c", "advice.detachedHead=false", "-C", j.sourceDir(), "checkout", j.sha()))

	logger.Println("git checkout result:", err)

	return err
}

func (j *githubJob) nix(subcmd string, args ...string) (*bytes.Buffer, error) {
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

func (j *githubJob) nixInstantiate() ([]string, error) {
	drvs := []string{}

	out, err := runCmd(exec.Command("nix-instantiate",
		"-I", "./nix",
		"-I", j.sourceDir(),
		"--argstr", "pname", j.pname(),
		j.ciNixPath(),
	))
	if err != nil {
		return drvs, err
	}

	scanner := bufio.NewScanner(out)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		if len(words) == 4 && strings.HasPrefix(words[3], "/nix/store") {
			drvs = append(drvs, words[3])
		}
	}

	return drvs, nil
}

func (j *githubJob) nixBuild() error {
	updateBuildStatus(j, "build")
	j.status("pending", "Nix Build...")

	output, err := j.nix(
		"build", "--out-link", j.resultLink(), "-f", j.ciNixPath())

	j.writeOutput(output)

	if err != nil {
		j.status("failure", err.Error())
		updateBuildStatus(j, "failure")
		return errors.New("Nix Failure: " + err.Error())
	}

	j.onSuccess()

	return nil
}

func (j *githubJob) writeOutput(output *bytes.Buffer) {
	content := output.Bytes()
	j.writeOutputToFile("nix_log", content)
	j.writeOutputToDB("nix_log", content)
}

func (j *githubJob) writeOutputToFile(baseName string, output []byte) {
	pathName := filepath.Join(j.buildDir(), baseName)
	err := ioutil.WriteFile(pathName, output, 0644)
	if err != nil {
		logger.Printf("Failed to create file %s: %s\n", pathName, err)
		return
	}
}

func (j *githubJob) writeOutputToDB(basename string, output []byte) {
	logger.Println(string(output))
	err := insertLog(j.buildID, basename, string(output))
	if err != nil {
		logger.Println("Failed writing log to DB:", err)
	}
}

func (j *githubJob) status(state, description string) {
	logger.Println(j.id()+":", state, description)
	go setGithubStatus(
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

	_, err := runCmd(exec.Command(
		"ssh", "root@3.120.166.103",
		"nix", "copy",
		"--all",
		"--to", "s3://scylla-cache?region=eu-central-1",
	))
	if err != nil {
		j.onError(err, "failed copying results to binary cache")
	}
}

func (j *githubJob) findOrCreateProjectID() (int, error) {
	return findOrCreateProjectID(j.Hook.Repository.FullName)
}

func (j *githubJob) createBuild(projectID int) (int, error) {
	return insertBuild(j.conn, projectID, j)
}

var sanitizeUrlPath = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

func (j *githubJob) saneFullName() string {
	return sanitizeUrlPath.ReplaceAllString(j.Hook.Repository.FullName, "_")
}

func (j *githubJob) targetURL() string {
	uri, _ := url.Parse(j.Host)
	uri.Path = "/builds/" + j.fullName() + "/" + j.sha()
	return uri.String()
}

func (j *githubJob) id() string         { return j.cloneURL() + "/" + j.sha() }
func (j *githubJob) fullName() string   { return j.Hook.Repository.FullName }
func (j *githubJob) cloneURL() string   { return j.Hook.PullRequest.Head.Repo.CloneURL }
func (j *githubJob) sha() string        { return j.Hook.PullRequest.Head.Sha }
func (j *githubJob) pname() string      { return j.saneFullName() + "-" + j.sha() }
func (j *githubJob) rootDir() string    { return config.BuildDir }
func (j *githubJob) buildDir() string   { return cleanJoin(j.rootDir(), j.saneFullName(), j.sha()) }
func (j *githubJob) sourceDir() string  { return filepath.Join(j.buildDir(), "source") }
func (j *githubJob) resultLink() string { return filepath.Join(j.buildDir(), "result") }
func (j *githubJob) ciNixPath() string  { return filepath.Join(j.buildDir(), "source", "ci.nix") }

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

	err := json.NewEncoder(body).Encode(&status)
	if err != nil {
		logger.Fatalf("Failed marshaling Github status: %s\n", err)
	}

	req, err := http.NewRequest("POST", statusURL, body)
	if err != nil {
		logger.Fatalf("Failed creating request: %s\n", err)
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

func githubAuthKey(givenURL, token string) string {
	u, err := url.Parse(givenURL)
	if err != nil {
		logger.Fatalln("Couldn't parse github url", err)
	}
	u.User = url.UserPassword(token, "x-oauth-basic")
	return "url." + u.String() + ".insteadOf"
}
