package server

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx"
)

type logListener struct {
	buildID int64
	recv    chan *logLine
}

type logLine struct {
	ID      int64     `json:"id"`
	BuildID int64     `json:"build_id"`
	Time    time.Time `json:"created_at"`
	Line    string    `json:"line"`
}

var logListenerRegister chan *logListener
var logListenerUnregister chan *logListener

func init() {
	logListenerRegister = make(chan *logListener)
	logListenerUnregister = make(chan *logListener)
}

func startLogDistributor(pool *pgx.ConnPool) {
	listeners := map[int64]map[*logListener]bool{}
	distribution := make(chan *logLine, 1000)

	go listenLogs(pool, distribution)

	for {
		select {
		case listener := <-logListenerRegister:
			registered, ok := listeners[listener.buildID]
			if ok {
				registered[listener] = true
			} else {
				listeners[listener.buildID] = map[*logListener]bool{listener: true}
			}
		case listener := <-logListenerUnregister:
			unregisterLogListener(listeners, listener)
		case ll := <-distribution:
			if registered, ok := listeners[ll.BuildID]; ok {
				for listener := range registered {
					select {
					case listener.recv <- ll:
					default:
						// it doesn't accept messages anymore, must have died of unnatural causes
						unregisterLogListener(listeners, listener)
					}
				}
			}
		}
	}
}

func unregisterLogListener(listeners map[int64]map[*logListener]bool, listener *logListener) {
	registered, ok := listeners[listener.buildID]
	if ok {
		delete(registered, listener)
		close(listener.recv)
		if len(registered) == 0 {
			delete(listeners, listener.buildID)
		}
	}
}

// Instead of creating a connection for each websocket, we multiplex
// notifications using channels, since that's much more efficient and we can
// have only a limited amount of connections but potentially millions of
// goroutines at a time
func listenLogs(pool *pgx.ConnPool, distribution chan *logLine) {
	conn, err := pool.Acquire()
	defer func() {
		pool.Release(conn)
	}()
	if err != nil {
		logger.Println(err)
		return
	}

	err = conn.Listen("loglines")
	if err != nil {
		logger.Println(err)
		return
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
		noti, err := conn.WaitForNotification(ctx)
		cancel()
		if err != nil {
			logger.Println(err)
			continue
		}

		ll := logLine{}
		err = json.NewDecoder(bytes.NewBufferString(noti.Payload)).Decode(&ll)
		if err != nil {
			logger.Println(err)
		}
		distribution <- &ll
	}
}

func forwardLogToDB(conn *pgx.Conn, buildID int64, line string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := conn.ExecEx(ctx, `INSERT INTO loglines (build_id, line) VALUES ($1, $2);`, nil, buildID, line)
	if err != nil {
		logger.Println(err)
	}
}
