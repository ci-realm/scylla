package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/k0kubun/pp"
	_ "github.com/lib/pq"
	"gopkg.in/go-playground/webhooks.v5/github"
)

type pgxLogger struct{}

func (l pgxLogger) Log(lvl pgx.LogLevel, msg string, data map[string]interface{}) {
	_, _ = pp.Println(msg, data)
}

var pgxpool *pgx.ConnPool

func SetupDB() {
	pgxcfg, err := pgx.ParseURI(config.DatabaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	pgxcfg.LogLevel = pgx.LogLevelWarn
	pgxcfg.Logger = pgxLogger{}

	pgxpool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     pgxcfg,
		AfterConnect:   func(*pgx.Conn) error { return nil },
		MaxConnections: 20,
	})
	if err != nil {
		logger.Fatalln("Couldn't connect to database:", err)
	}
}

type dbProject struct {
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Name       string    `json:"name"`
	BuildCount int       `json:"buildCount"`
}

type dbBuild struct {
	ID          int64                      `json:"id"`
	Status      string                     `json:"status"`
	CreatedAt   time.Time                  `json:"createdAt"`
	UpdatedAt   time.Time                  `json:"updatedAt"`
	StatusAt    time.Time                  `json:"statusAt"`
	FinishedAt  time.Time                  `json:"finishedAt"`
	Hook        *github.PullRequestPayload `json:"hook"`
	ProjectName string                     `json:"projectName"`
	Log         []*logLine                 `json:"log"`
}

func (b dbBuild) BranchName() string       { return b.Hook.PullRequest.Head.Ref }
func (b dbBuild) Owner() string            { return b.Hook.Repository.Owner.Login }
func (b dbBuild) ProjectLink() string      { return "/builds/" + b.ProjectName }
func (b dbBuild) Repo() string             { return b.Hook.Repository.Name }
func (b dbBuild) Title() string            { return b.Hook.PullRequest.Title }
func (b dbBuild) GithubLink() string       { return b.Hook.PullRequest.HTMLURL }
func (b dbBuild) SHA() string              { return b.Hook.PullRequest.Head.Sha }
func (b dbBuild) BuildTime() time.Duration { return b.FinishedAt.Sub(b.CreatedAt) }
func (b dbBuild) CommitLink() string {
	return b.Hook.PullRequest.Base.Repo.HTMLURL + "/commit/" + b.Hook.PullRequest.Head.Sha
}

func (b dbBuild) BranchLink() string {
	return b.Hook.PullRequest.Base.Repo.HTMLURL + "/tree/" + b.Hook.PullRequest.Head.Ref
}
func (b dbBuild) BuildLink() string {
	return fmt.Sprintf("/builds/%s/%s/%d", b.Owner(), b.Repo(), b.ID)
}
func (b dbBuild) RestartLink() string {
	return fmt.Sprintf("/builds/%s/%s/%d/restart", b.Owner(), b.Repo(), b.ID)
}

func insertBuild(db *pgx.Conn, projectID int, job *githubJob) (int, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(job.Hook); err != nil {
		return 0, err
	}

	var buildID int
	err := db.QueryRow(
		`INSERT INTO builds (project_id, data) VALUES ($1, $2) RETURNING id;`,
		projectID, buf.String()).Scan(&buildID)

	return buildID, err
}

func findBuildByID(db *pgx.Conn, buildID int) (*githubJob, error) {
	var projectID int
	var rawData []byte
	err := db.QueryRow(
		`SELECT project_id, data FROM builds WHERE id = $1;`,
		buildID).Scan(&projectID, &rawData)
	if err != nil {
		return nil, err
	}

	hook := &github.PullRequestPayload{}
	err = json.NewDecoder(bytes.NewBuffer(rawData)).Decode(hook)
	if err != nil {
		return nil, err
	}
	return &githubJob{Hook: hook, conn: db, buildID: buildID}, nil
}

type dbOrg struct {
	Owner      string `json:"owner"`
	URL        string `json:"url"`
	BuildCount int64  `json:"buildCount"`
}

func findOrganizations(db *pgx.Conn) ([]dbOrg, error) {
	orgs := []dbOrg{}
	rows, err := db.Query(
		`SELECT
      data#>>'{pull_request, head, repo, owner, login}' AS owner,
      data#>>'{pull_request, head, repo, owner, html_url}' AS url,
      count(id)
      FROM builds
      GROUP BY url, owner`,
	)

	if err != nil {
		return orgs, err
	}

	for rows.Next() {
		org := dbOrg{}
		err = rows.Scan(&org.Owner, &org.URL, &org.BuildCount)
		if err != nil {
			return orgs, err
		}
		orgs = append(orgs, org)
	}

	return orgs, err
}

// TODO: improve performance by reducing the builds.data
func findBuilds(db *pgx.Conn, orgName string) ([]dbBuild, error) {
	builds := []dbBuild{}
	var rows *pgx.Rows
	var err error

	if orgName == "" {
		rows, err = db.Query(
			`SELECT
         builds.id,
         builds.status,
         builds.created_at,
         builds.updated_at,
         builds.status_at,
         builds.finished_at,
         projects.name,
         builds.data
       FROM builds
       JOIN projects ON projects.id = builds.project_id
       ORDER BY builds.created_at DESC
       LIMIT 100;`,
		)
	} else {
		rows, err = db.Query(
			`SELECT
         builds.id,
         builds.status,
         builds.created_at,
         builds.updated_at,
         builds.status_at,
         builds.finished_at,
         projects.name,
         builds.data
       FROM builds
       JOIN projects ON projects.id = builds.project_id
       WHERE data#>>'{pull_request, head, repo, owner, login}' = $1
       ORDER BY builds.created_at DESC
       LIMIT 100;`,
			orgName,
		)
	}

	if err != nil {
		return builds, err
	}

	for rows.Next() {
		createdAt := &pgtype.Timestamptz{}
		updatedAt := &pgtype.Timestamptz{}
		statusAt := &pgtype.Timestamptz{}
		finishedAt := &pgtype.Timestamptz{}
		build := dbBuild{Hook: &github.PullRequestPayload{}}

		var buildData []byte

		err = rows.Scan(
			&build.ID,
			&build.Status,
			createdAt,
			updatedAt,
			statusAt,
			finishedAt,
			&build.ProjectName,
			&buildData,
		)
		if err != nil {
			return builds, err
		}

		build.CreatedAt = createdAt.Time
		build.UpdatedAt = updatedAt.Time
		build.StatusAt = statusAt.Time
		build.FinishedAt = finishedAt.Time

		err = json.NewDecoder(bytes.NewBuffer(buildData)).Decode(&build.Hook)
		builds = append(builds, build)
		if err != nil {
			return builds, err
		}
	}

	return builds, nil
}

func findFullBuildByID(db *pgx.Conn, buildID int64) (*dbBuild, error) {
	var buildData []byte
	createdAt := &pgtype.Timestamptz{}
	updatedAt := &pgtype.Timestamptz{}
	finishedAt := &pgtype.Timestamptz{}
	build := &dbBuild{Hook: &github.PullRequestPayload{}}

	logLines := &pgtype.TextArray{}
	logTimes := &pgtype.TimestampArray{}
	logContent := pgtype.Text{}

	err := db.QueryRow(
		`SELECT
       builds.id,
       builds.status,
       builds.created_at,
       builds.updated_at,
       builds.finished_at,
       builds.data,
       array_agg(loglines.line order by loglines.id),
       array_agg(loglines.created_at order by loglines.id),
       (SELECT content FROM logs WHERE logs.build_id = $1),
       projects.name
     FROM builds
     JOIN projects ON projects.id = builds.project_id
     LEFT OUTER JOIN loglines ON loglines.build_id = builds.id
     LEFT OUTER JOIN logs ON logs.build_id = builds.project_id
     WHERE builds.id = $1
     GROUP BY projects.id, builds.id;`,
		buildID,
	).Scan(
		&build.ID,
		&build.Status,
		createdAt,
		updatedAt,
		finishedAt,
		&buildData,
		logLines,
		logTimes,
		&logContent,
		&build.ProjectName,
	)

	build.CreatedAt = createdAt.Time
	build.UpdatedAt = updatedAt.Time
	build.FinishedAt = finishedAt.Time

	if len(logLines.Elements) > 0 {
		build.Log = make([]*logLine, len(logLines.Elements))
		for n, line := range logLines.Elements {
			build.Log[n] = &logLine{Time: logTimes.Elements[n].Time, Line: line.String}
		}
	}

	if logContent.String != "" {
		build.Log = []*logLine{}
		for _, line := range strings.Split(logContent.String, "\n") {
			build.Log = append(build.Log, &logLine{Line: line})
		}
	}

	// "("2018-10-14 12:21:23.827167+00","Line 46")"

	if err != nil {
		return build, err
	}

	err = json.NewDecoder(bytes.NewBuffer(buildData)).Decode(&build.Hook)
	return build, err
}

func (d dbProject) Link() string {
	return "/builds/" + d.Name
}

func updateBuildStatus(job *githubJob, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	tx, err := pgxpool.BeginEx(ctx, nil)
	defer func() { cancel(); _ = tx.Rollback() }()
	if err != nil {
		logger.Println("Failed updating build status:", err)
		return
	}

	_, err = tx.ExecEx(ctx, `SET idle_in_transaction_session_timeout TO '1000';`, nil)
	if err != nil {
		logger.Println("Failed updating build status:", err)
		return
	}

	_, err = tx.ExecEx(ctx, `UPDATE builds SET status = $1, status_at = now() WHERE id = $2;`, nil, status, job.buildID)
	if err != nil {
		logger.Println("Failed updating build status:", err)
		return
	}

	err = tx.CommitEx(ctx)
	if err != nil {
		logger.Println("Failed updating build status:", err)
	}
}

func findOrCreateProjectID(name string) (int, error) {
	var projectID int
	err := pgxpool.QueryRow(
		`INSERT INTO projects (name, created_at) VALUES ($1, $2)
       ON CONFLICT (name) DO
         UPDATE SET name = $1
     RETURNING id;`,
		name, time.Now().UTC(),
	).Scan(&projectID)
	logger.Println("projectID:", projectID, err)

	return projectID, err
}

func insertLog(buildID int, kind, content string) error {
	_, err := pgxpool.Exec(
		`INSERT INTO logs (build_id, kind, content) VALUES ($1, $2, $3)`,
		buildID, kind, content)
	return err
}

func insertResult(buildID int, path string) error {
	_, err := pgxpool.Exec(
		`INSERT INTO results (build_id, path) VALUES ($1, $2)`,
		buildID, path)
	return err
}

func compactLog(buildID int) error {
	tx, err := pgxpool.Begin()
	if err != nil {
		return fmt.Errorf("Failed starting transaction: %s", err)
	}

	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(
		`INSERT INTO logs (build_id, content) SELECT $1, string_agg(created_at::text || ' ' || line, '\n')
     FROM loglines
     WHERE build_id = $1`,
		buildID,
	)
	if err != nil {
		return fmt.Errorf("Failed inserting logs: %s", err)
	}

	_, err = tx.Exec(
		`DELETE FROM loglines WHERE build_id = $1`,
		buildID,
	)
	if err != nil {
		return fmt.Errorf("Failed deleting old logs: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Failed transaction for compactLog: %s", err)
	}
	return nil
}
