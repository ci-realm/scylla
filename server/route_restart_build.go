package server

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/manveru/scylla/queue"
)

func postBuildsProjectIdRestart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`Invalid project id, must be integer`))
		return
	}

	err = withConn(func(conn *pgx.Conn) error {
		build, err := findFullBuildByID(conn, buildID)
		if err != nil {
			return err
		}

		item := &queue.Item{Args: map[string]interface{}{"build_id": buildID, "Host": progressHost(r)}}
		err = jobQueue.Insert(item)
		if err != nil {
			return err
		}

		w.Header().Set("Location", build.BuildLink())
		w.WriteHeader(302)
		return nil
	})

	if err != nil {
		logger.Println(err)
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}
