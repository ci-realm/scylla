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
	r.HandleFunc("/socket", upgradeWebSocket).Methods("GET")

	if config.Mode == "development" {
		r.PathPrefix("/").HandlerFunc(getIndex)
	}

	// r.Get("/socket", sockets.JSON(Message{}, &sockets.Options{
	// 	Logger:            logger,
	// 	LogLevel:          sockets.LogLevelWarning,
	// 	SkipLogging:       false,
	// 	WriteWait:         60 * time.Second,
	// 	PongWait:          60 * time.Second,
	// 	PingPeriod:        (60 * time.Second * 8 / 10),
	// 	MaxMessageSize:    65536,
	// 	SendChannelBuffer: 10,
	// 	RecvChannelBuffer: 10,
	// 	// AllowedOrigin:     "https?://{{host}}$",
	// 	AllowedOrigin: ".",
	// }), handleWebSocket)
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
