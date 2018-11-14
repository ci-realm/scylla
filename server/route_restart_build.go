package server

import (
	"github.com/jackc/pgx"
	"github.com/manveru/scylla/queue"
	macaron "gopkg.in/macaron.v1"
)

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
