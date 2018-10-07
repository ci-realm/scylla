package main

import (
	"github.com/jackc/pgx"
	"github.com/manveru/scylla/queue"
	macaron "gopkg.in/macaron.v1"
)

func getBuildsProjectId(ctx *macaron.Context) {
	projectName := ctx.Params("user") + "/" + ctx.Params("repo")
	buildID := ctx.ParamsInt("id")

	withConn(ctx, func(conn *pgx.Conn) error {
		build, err := findBuildByProjectAndID(conn, projectName, buildID)
		if err != nil {
			return err
		}

		ctx.Data["Build"] = build
		ctx.HTML(200, "builds_project_id")
		return nil
	})
}

func postBuildsProjectIdRestart(ctx *macaron.Context) {
	projectName := ctx.Params("user") + "/" + ctx.Params("repo")
	buildID := ctx.ParamsInt("id")

	withConn(ctx, func(conn *pgx.Conn) error {
		build, err := findBuildByProjectAndID(conn, projectName, buildID)
		if err != nil {
			return err
		}

		item := &queue.Item{Args: map[string]interface{}{"build_id": buildID, "Host": progressHost(ctx)}}
		err = jobQueue.Insert(item)
		if err != nil {
			return err
		}

		ctx.Redirect(build.BuildLink(), 302)
		return nil
	})
}
