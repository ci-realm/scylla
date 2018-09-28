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

func withConn(ctx *macaron.Context, f func(*pgx.Conn) error) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		ctx.Error(500, err.Error())
		return
	}
	defer pgxpool.Release(conn)
	err = f(conn)
	if err != nil {
		ctx.Error(500, err.Error())
	}
}
