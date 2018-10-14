package server

import (
	"github.com/jackc/pgx"
	macaron "gopkg.in/macaron.v1"
)

func getDashboardBuilds(ctx *macaron.Context) {
	// projectName := ctx.Params("user") + "/" + ctx.Params("repo")

	withConn(ctx, func(conn *pgx.Conn) error {
		builds, err := findBuilds(conn)
		if err != nil {
			return err
		}
		ctx.Data["Builds"] = builds
		ctx.HTML(200, "dashboard_builds")
		return nil
	})
}
