package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/manveru/scylla/queue"
)

func handleWebSocket(r *http.Request, conn *websocket.Conn) {
	socket := &webSocket{conn: conn, host: progressHost(r)}
	socket.mainLoop()
}

type webSocket struct {
	conn     *websocket.Conn
	host     string
	listener *logListener
}

func (s *webSocket) mainLoop() {
	for {
		msg := &Message{}
		err := s.conn.ReadJSON(msg)
		if err != nil {
			log.Println(err)
			return
		}

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
		}
	}
}

func (s *webSocket) writeData(mutation string, data msgData) {
	if err := s.conn.WriteJSON(&Message{
		Mutation: mutation,
		Data:     data,
	}); err != nil {
		logger.Println(err)
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

	if s.listener != nil {
		logger.Println("Unregister from existing")
		logListenerUnregister <- s.listener
		s.listener = nil
	}

	recv := make(chan *logLine)
	listener := &logListener{buildID: id, recv: recv}
	logListenerRegister <- listener

	s.listener = listener

	go func() {
		for line := range recv {
			s.writeData("buildLog", msgData{
				"time": line.Time,
				"line": line.Line,
			})
		}
	}()
}

func (s *webSocket) getBuildLogUnwatch(data msgData) {
	if s.listener != nil {
		logger.Println("Unregister from build log")
		logListenerUnregister <- s.listener
		s.listener = nil
	}
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
