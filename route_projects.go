package main

import (
	"github.com/jackc/pgx"
	macaron "gopkg.in/macaron.v1"
)

func getProjects(ctx *macaron.Context) {
	withConn(ctx, func(conn *pgx.Conn) error {
		projects, err := findAllProjects(conn, 100)
		if err != nil {
			return err
		}

		ctx.Data["Projects"] = projects
		ctx.HTML(200, "projects")
		return nil
	})
}
