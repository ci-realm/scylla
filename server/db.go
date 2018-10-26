package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/k0kubun/pp"
	_ "github.com/lib/pq"
)

type pgxLogger struct{}

func (l pgxLogger) Log(lvl pgx.LogLevel, msg string, data map[string]interface{}) {
	pp.Println(msg, data)
}

var pgxpool *pgx.ConnPool

func SetupDB() {
	pgxcfg, err := pgx.ParseURI(config.DatabaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	pgxcfg.LogLevel = pgx.LogLevelWarn
	pgxcfg.Logger = pgxLogger{}

	if os.Getenv("HOST") != "" && strings.Contains(config.DatabaseURL, "amazonaws.com") {
		tunnelStarted := make(chan bool)
		go setupDatabaseTunnel(tunnelStarted)
		<-tunnelStarted
		pgxcfg.Port = localDBPort
		pgxcfg.Host = "127.0.0.1"
	}

	pgxpool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     pgxcfg,
		AfterConnect:   func(*pgx.Conn) error { return nil },
		MaxConnections: 20,
	})
	if err != nil {
		logger.Fatalln(err)
	}

	// go streamFakeLog(1)
}

func streamFakeLog(buildID int64) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Fatal(err)
	}

	n := 0
	for {
		n++
		forwardLogToDB(conn, buildID, fmt.Sprintf("Line %d", n))

		time.Sleep(time.Second)
	}
}

type dbProject struct {
	ID         int64
	CreatedAt  *pgtype.Timestamptz
	UpdatedAt  *pgtype.Timestamptz
	Name       string
	BuildCount int
}

type dbBuild struct {
	ID          int64
	Status      string
	CreatedAt   *pgtype.Timestamptz
	UpdatedAt   *pgtype.Timestamptz
	StatusAt    *pgtype.Timestamptz
	FinishedAt  *pgtype.Timestamptz
	Hook        GithubHook
	ProjectName string
	Log         []*logLine
}

func (b dbBuild) BranchName() string       { return b.Hook.PullRequest.Head.Ref }
func (b dbBuild) Owner() string            { return b.Hook.Repository.Owner.Login }
func (b dbBuild) ProjectLink() string      { return "/builds/" + b.ProjectName }
func (b dbBuild) Repo() string             { return b.Hook.Repository.Name }
func (b dbBuild) Title() string            { return b.Hook.PullRequest.Title }
func (b dbBuild) GithubLink() string       { return b.Hook.PullRequest.HTMLURL }
func (b dbBuild) SHA() string              { return b.Hook.PullRequest.Head.Sha }
func (b dbBuild) BuildTime() time.Duration { return b.FinishedAt.Time.Sub(b.CreatedAt.Time) }
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

	hook := &GithubHook{}
	err = json.NewDecoder(bytes.NewBuffer(rawData)).Decode(hook)
	if err != nil {
		return nil, err
	}
	return &githubJob{Hook: hook, conn: db, buildID: buildID}, nil
}

type dbOrg struct {
	Owner      string
	URL        string
	BuildCount int64
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
		build := dbBuild{Hook: GithubHook{},
			CreatedAt:  &pgtype.Timestamptz{},
			UpdatedAt:  &pgtype.Timestamptz{},
			StatusAt:   &pgtype.Timestamptz{},
			FinishedAt: &pgtype.Timestamptz{},
		}

		var buildData []byte

		err = rows.Scan(
			&build.ID,
			&build.Status,
			build.CreatedAt,
			build.UpdatedAt,
			build.StatusAt,
			build.FinishedAt,
			&build.ProjectName,
			&buildData,
		)
		if err != nil {
			return builds, err
		}

		err = json.NewDecoder(bytes.NewBuffer(buildData)).Decode(&build.Hook)
		builds = append(builds, build)
		if err != nil {
			return builds, err
		}
	}

	return builds, nil
}

func findBuildByProjectAndID(db *pgx.Conn, projectName string, buildID int) (*dbBuild, error) {
	var buildData []byte
	build := &dbBuild{Hook: GithubHook{},
		CreatedAt:   &pgtype.Timestamptz{},
		UpdatedAt:   &pgtype.Timestamptz{},
		FinishedAt:  &pgtype.Timestamptz{},
		ProjectName: projectName,
	}

	logLines := &pgtype.TextArray{}
	logTimes := &pgtype.TimestampArray{}

	err := db.QueryRow(
		`SELECT
       builds.id,
       builds.status,
       builds.created_at,
       builds.updated_at,
       builds.finished_at,
       builds.data,
       array_agg(loglines.line order by loglines.id),
       array_agg(loglines.created_at order by loglines.id)
     FROM builds
     LEFT OUTER JOIN loglines ON loglines.build_id = builds.id
     WHERE builds.id = $1
     GROUP BY builds.id;`,
		buildID,
	).Scan(
		&build.ID,
		&build.Status,
		build.CreatedAt,
		build.UpdatedAt,
		build.FinishedAt,
		&buildData,
		logLines,
		logTimes,
	)

	build.Log = make([]*logLine, len(logLines.Elements))
	for n, line := range logLines.Elements {
		build.Log[n] = &logLine{Time: logTimes.Elements[n].Time, Line: line.String}
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

func findBuildsByProjectName(db *pgx.Conn, projectName string) ([]dbBuild, error) {
	rows, err := db.Query(
		`SELECT builds.id, builds.status, builds.created_at, builds.updated_at, builds.data FROM builds
     JOIN projects on projects.id = builds.project_id
     WHERE projects.name = $1
     ORDER BY builds.created_at DESC LIMIT 100;`,
		projectName,
	)

	builds := []dbBuild{}
	for rows.Next() {
		build := dbBuild{Hook: GithubHook{},
			CreatedAt:   &pgtype.Timestamptz{},
			UpdatedAt:   &pgtype.Timestamptz{},
			ProjectName: projectName,
		}
		var buildData []byte

		err := rows.Scan(&build.ID, &build.Status, build.CreatedAt, build.UpdatedAt, &buildData)
		if err != nil {
			return nil, err
		}

		err = json.NewDecoder(bytes.NewBuffer(buildData)).Decode(&build.Hook)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, err
}

func findAllProjects(db *pgx.Conn, limit int) ([]dbProject, error) {
	rows, err := db.Query(
		`SELECT projects.id, projects.name, projects.created_at, count(distinct(builds.id)) FROM projects
     JOIN builds ON builds.project_id = projects.id
     GROUP BY projects.id
     LIMIT $1;`,
		limit,
	)
	if err != nil {
		return nil, err
	}

	out := []dbProject{}
	for rows.Next() {
		project := dbProject{CreatedAt: &pgtype.Timestamptz{}}
		err := rows.Scan(&project.ID, &project.Name, project.CreatedAt, &project.BuildCount)
		if err != nil {
			return nil, err
		}
		out = append(out, project)
	}
	return out, err
}

func updateBuildStatus(job *githubJob, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	tx, err := pgxpool.BeginEx(ctx, nil)
	if err != nil {
		logger.Println("Failed updating build status:", err)
	}
	defer func() { cancel(); _ = tx.Rollback() }()

	_, err = tx.ExecEx(ctx, `SET idle_in_transaction_session_timeout TO '1000';`, nil)
	if err != nil {
		logger.Println("Failed updating build status:", err)
	}

	_, err = tx.ExecEx(ctx, `UPDATE builds SET status = $1, status_at = now() WHERE id = $2;`, nil, status, job.buildID)
	if err != nil {
		logger.Println("Failed updating build status:", err)
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
