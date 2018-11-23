package server

import (
	"time"

	"github.com/go-macaron/sockets"
	"github.com/jackc/pgx"
	macaron "gopkg.in/macaron.v1"
)

func setupRouting(m *macaron.Macaron) {
	if config.Mode == "development" {
		m.Any("/*", getIndex)
	}

	m.Get("/_system/alive", getAlive)
	m.Post("/builds/:user/:repo/:id/restart", postBuildsProjectIdRestart)

	m.Post("/hooks/github", postHooksGithub)
	m.Get("/socket", sockets.JSON(Message{}, &sockets.Options{
		Logger:            logger,
		LogLevel:          sockets.LogLevelWarning,
		SkipLogging:       false,
		WriteWait:         60 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        (60 * time.Second * 8 / 10),
		MaxMessageSize:    65536,
		SendChannelBuffer: 10,
		RecvChannelBuffer: 10,
		// AllowedOrigin:     "https?://{{host}}$",
		AllowedOrigin: ".",
	}), handleWebSocket)
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
