package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx"
)

func setupRouting(r *mux.Router) {
	r.HandleFunc("/_system/alive", getAlive).Methods("GET")
	r.HandleFunc("/builds/{id}/restart", postBuildsProjectIdRestart).Methods("GET")
	r.HandleFunc("/hooks/github", postHooksGithub).Methods("POST")
	// r.HandleFunc("/hooks/gitlab", postHooksGitlab).Methods("POST")
	r.HandleFunc("/socket", upgradeWebSocket).Methods("GET")

	if config.Mode == "development" {
		r.PathPrefix("/").HandlerFunc(proxyIndex)
	} else {
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
		r.PathPrefix("/").HandlerFunc(getIndex)
	}
}

var webSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(req *http.Request) bool { return true },
}

func upgradeWebSocket(wri http.ResponseWriter, req *http.Request) {
	conn, err := webSocketUpgrader.Upgrade(wri, req, nil)
	if err != nil {
		logger.Println(err)
		return
	}
	handleWebSocket(req, conn)
}

func withConn(f func(*pgx.Conn) error) error {
	conn, err := pgxpool.Acquire()
	if err != nil {
		return err
	}
	defer pgxpool.Release(conn)
	return f(conn)
}
