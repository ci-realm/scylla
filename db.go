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
	Hook        GithubHook
	ProjectName string
	Stdout      string
	Stderr      string
}

func (b dbBuild) ProjectLink() string {
	return "/builds/" + b.ProjectName
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

func findBuildByProjectAndID(db *pgx.Conn, projectName string, buildID int) (dbBuild, error) {
	var buildData []byte
	build := dbBuild{Hook: GithubHook{},
		CreatedAt:   &pgtype.Timestamptz{},
		UpdatedAt:   &pgtype.Timestamptz{},
		ProjectName: projectName,
	}

	err := db.QueryRow(
		`SELECT
        builds.id,
        builds.status,
        builds.created_at,
        builds.updated_at,
        builds.data,
        stderr.content,
        stdout.content
      FROM builds
      JOIN projects ON projects.id = builds.project_id
      JOIN logs AS stderr ON stderr.build_id = builds.id AND stderr.kind = 'stderr'
      JOIN logs AS stdout ON stdout.build_id = builds.id AND stdout.kind = 'stdout'
      WHERE projects.name = $1 AND builds.id = $2;`,
		projectName, buildID,
	).Scan(
		&build.ID,
		&build.Status,
		build.CreatedAt,
		build.UpdatedAt,
		&buildData,
		&build.Stderr,
		&build.Stdout,
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
     WHERE projects.name = $1;`,
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
	_, err := pgxpool.Exec(`UPDATE builds SET status = $1 WHERE id = $2;`, status, job.buildID)
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
