package main

import (
	"github.com/jackc/pgx"
	macaron "gopkg.in/macaron.v1"
)

func getBuildsProject(ctx *macaron.Context) {
	projectName := ctx.Params("user") + "/" + ctx.Params("repo")

	withConn(ctx, func(conn *pgx.Conn) error {
		builds, err := findBuildsByProjectName(conn, projectName)
		if err != nil {
			return err
		}
		ctx.Data["Builds"] = builds
		ctx.Data["Name"] = projectName
		ctx.HTML(200, "builds_project")
		return nil
	})
}
