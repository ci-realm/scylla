package main

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	_ "github.com/lib/pq"
)

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
	Log         *pgtype.Text
}

func (b dbBuild) BranchName() string       { return b.Hook.PullRequest.Base.Ref }
func (b dbBuild) Owner() string            { return b.Hook.Repository.Owner.Login }
func (b dbBuild) ProjectLink() string      { return "/builds/" + b.ProjectName }
func (b dbBuild) Repo() string             { return b.Hook.Repository.Name }
func (b dbBuild) Title() string            { return b.Hook.PullRequest.Title }
func (b dbBuild) GithubLink() string       { return b.Hook.PullRequest.HTMLURL }
func (b dbBuild) SHA() string              { return b.Hook.PullRequest.Head.Sha }
func (b dbBuild) BuildTime() time.Duration { return b.FinishedAt.Time.Sub(b.CreatedAt.Time) }

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

func findBuilds(db *pgx.Conn) ([]dbBuild, error) {
	builds := []dbBuild{}
	rows, err := db.Query(
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
      LIMIT 100;`,
	)

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

		rows.Scan(
			&build.ID,
			&build.Status,
			build.CreatedAt,
			build.UpdatedAt,
			build.StatusAt,
			build.FinishedAt,
			&build.ProjectName,
			&buildData,
		)
		err = json.NewDecoder(bytes.NewBuffer(buildData)).Decode(&build.Hook)
		builds = append(builds, build)
		if err != nil {
			return builds, err
		}
	}

	return builds, nil
}

func findBuildByProjectAndID(db *pgx.Conn, projectName string, buildID int) (dbBuild, error) {
	var buildData []byte
	build := dbBuild{Hook: GithubHook{},
		CreatedAt:   &pgtype.Timestamptz{},
		UpdatedAt:   &pgtype.Timestamptz{},
		ProjectName: projectName,
		Log:         &pgtype.Text{},
	}

	err := db.QueryRow(
		`SELECT
        builds.id,
        builds.status,
        builds.created_at,
        builds.updated_at,
        builds.data,
        logs.content
				data#>>'{pull_request,head,repo,owner,login}' AS owner,
				data#>>'{pull_request,head,repo,name}' AS repo
			FROM builds
      JOIN projects ON projects.id = builds.project_id
      LEFT OUTER JOIN logs ON logs.build_id = builds.id
      WHERE projects.name = $1 AND builds.id = $2;`,
		projectName, buildID,
	).Scan(
		&build.ID,
		&build.Status,
		build.CreatedAt,
		build.UpdatedAt,
		&buildData,
		&build.Log,
	)

	if err != nil {
		return build, err
	}

	err = json.NewDecoder(bytes.NewBuffer(buildData)).Decode(&build.Hook)
	return build, err
}

func (d dbProject) Link() string {
	return "/builds/" + d.Name
}

func findProjectByID(db *pgx.Conn, projectID int) (dbProject, error) {
	project := dbProject{}
	err := db.QueryRow(
		`SELECT projects.id, projects.name, projects.created_at, projects.updated_at, count(distinct(builds.id))
     FROM projects
     JOIN builds on builds.project_id = projects.id
     WHERE id = $1
     GROUP BY projects.id;`,
		projectID,
	).Scan(&project.ID, &project.Name, &project.BuildCount)
	return project, err
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
	_, err := pgxpool.Exec(
		`UPDATE builds SET status = $1, status_at = now()
     WHERE id = $2;`,
		status, job.buildID)
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
