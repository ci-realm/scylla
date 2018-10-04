package main

import (
	"github.com/jackc/pgx"
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
