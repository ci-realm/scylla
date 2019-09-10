package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/manveru/scylla/queue"
)

const (
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Empty the messages outbox every period.
	msgsPeriod = 1 * time.Second

	// Time allowed to write the message to the client.
	writeWait = 10 * time.Second
)

func handleWebSocket(r *http.Request, conn *websocket.Conn) {
	conn.SetReadLimit(512)

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	socket := &webSocket{
		conn:   conn,
		host:   progressHost(r),
		outbox: make(chan *Message),
	}

	go socket.writer()
	socket.reader()
}

type webSocket struct {
	conn     *websocket.Conn
	host     string
	listener *logListener
	outbox   chan *Message
}

func (s *webSocket) writer() {
	pingTicker := time.NewTicker(pingPeriod)
	msgsTicker := time.NewTicker(msgsPeriod)

	defer func() {
		pingTicker.Stop()
		msgsTicker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case msg := <-s.outbox:
			if msg != nil {
				s.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := s.conn.WriteJSON(msg); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
        logger.Println(err)
				return
			}
		}
	}
}

func (s *webSocket) reader() {
	defer s.conn.Close()
	s.conn.SetReadLimit(512)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msg := &Message{}
		err := s.conn.ReadJSON(msg)
		if err != nil {
			logger.Println(err)
			s.cleanup()
			return
		}

		s.handleMessage(msg)
	}
}

func (s *webSocket) handleMessage(msg *Message) {
	logger.Println(msg.Kind)

	switch msg.Kind {
	case "restart":
		s.restartBuild(msg.Data)
	case "lastBuilds":
		s.getLastBuilds(msg.Data)
	case "organizations":
		s.getOrganizations(msg.Data)
	case "organizationBuilds":
		s.getOrganizationBuilds(msg.Data)
	case "build":
		s.getBuild(msg.Data)
	case "buildLogWatch":
		s.getBuildLogWatch(msg.Data)
	case "buildLogUnwatch":
		s.getBuildLogUnwatch(msg.Data)
	default:
		s.wsError(fmt.Errorf("Unknown message: %v", msg.Kind))
	}
}

func (s *webSocket) writeData(mutation string, data msgData) {
	s.outbox <- &Message{
		Mutation: mutation,
		Data:     data,
	}
}

func (s *webSocket) getOrganizations(data msgData) {
	s.writeData("organizations", msgData{
		"organizations": wsOrganizations(),
	})
}

func (s *webSocket) getOrganizationBuilds(data msgData) {
	orgName, err := data.getString("orgName")
	if err != nil {
		logger.Println(err)

		s.wsError(err)
		return
	}

	s.writeData("organizationBuilds", msgData{
		"organizationBuilds": wsLatestBuildsForOrg(orgName),
	})
}

func (s *webSocket) restartBuild(d msgData) {
	buildID, err := d.getInt64("id")
	if err != nil {
		logger.Println(err)
	}
	item := &queue.Item{Args: msgData{"build_id": buildID, "Host": s.host}}
	err = jobQueue.Insert(item)
	if err != nil {
		logger.Println(err)
	}
	s.writeData("restart", msgData{})
}

func (s *webSocket) getLastBuilds(_ msgData) {
	s.writeData("lastBuilds", msgData{"builds": wsLatestBuilds()})
}

func (s *webSocket) getBuild(data msgData) {
	id, err := data.getID()
	if err != nil {
		logger.Println("parsing id", err)
		return
	}

	build := wsBuild(id)
	if build != nil {
		s.writeData("build", msgData{"build": build})
	}
}

func (s *webSocket) getBuildLogWatch(data msgData) {
	id, err := data.getID()
	if err != nil {
		return
	}

	s.cleanup()

	recv := make(chan *logLine)
	listener := &logListener{buildID: id, recv: recv}
	logListenerRegister <- listener

	s.listener = listener

	go func() {
		for line := range recv {
			s.writeData("buildLog", msgData{
        "buildId": line.BuildID,
				"createdAt": line.Time,
				"line": line.Line,
			})
		}
	}()
}

func (s *webSocket) getBuildLogUnwatch(data msgData) {
	s.cleanup()
}

func wsOrganizations() (orgs []dbOrg) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Panic(err)
		return
	}
	defer pgxpool.Release(conn)

	orgs, err = findOrganizations(conn)

	if err != nil {
		logger.Panic(err)
	}
	return
}

func wsLatestBuilds() (builds []dbBuild) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Panic(err)
		return
	}
	defer pgxpool.Release(conn)

	builds, err = findBuilds(conn, "")

	if err != nil {
		logger.Panic(err)
	}

	return
}

func wsLatestBuildsForOrg(orgName string) (builds []dbBuild) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Panic(err)
		return
	}
	defer pgxpool.Release(conn)

	builds, err = findBuilds(conn, orgName)

	if err != nil {
		logger.Println(orgName, err)
	}

	return
}

func wsBuild(buildID int64) (build *dbBuild) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Panic(err)
		return
	}
	defer pgxpool.Release(conn)

	build, err = findFullBuildByID(conn, buildID)
	if err != nil {
		logger.Println(buildID, err)
	}

	return
}

func (s *webSocket) cleanup() {
	if s.listener != nil {
		logListenerUnregister <- s.listener
		s.listener = nil
	}
}

func (s *webSocket) wsError(err error) {
	s.writeData("error", msgData{
		"error": err,
	})
}
