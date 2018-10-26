package server

import (
	"strconv"

	macaron "gopkg.in/macaron.v1"
)

type webSocket struct {
	ctx          *macaron.Context
	receiver     <-chan *Message
	sender       chan<- *Message
	done         <-chan bool
	disconnect   chan<- int
	errorChannel <-chan error

	listener *logListener
}

func (s *webSocket) mainLoop() {
	for {
		select {
		case msg := <-s.receiver:
			logger.Println(msg.Kind)
			switch msg.Kind {
			case "last-builds":
				s.getLastBuilds(msg.Data)
			case "organizations":
				s.getOrganizations(msg.Data)
			case "organization-builds":
				s.getOrganizationBuilds(msg.Data)
			case "build":
				s.getBuild(msg.Data)
			case "build-log-watch":
				s.getBuildLogWatch(msg.Data)
			case "build-log-unwatch":
				s.getBuildLogUnwatch(msg.Data)
			}
		case <-s.done:
			return
		case err := <-s.errorChannel:
			logger.Println(err)
		}
	}
}

func (s *webSocket) getOrganizations(data map[string]interface{}) {
	s.sender <- &Message{
		Mutation: "ORGANIZATIONS",
		Data:     map[string]interface{}{"organizations": wsOrganizations()},
	}
}

func (s *webSocket) getOrganizationBuilds(data map[string]interface{}) {
	orgName, ok := data["orgName"].(string)
	if !ok {
		logger.Println("missing orgName")
		return
	}

	s.sender <- &Message{
		Mutation: "ORGANIZATION_BUILDS",
		Data:     map[string]interface{}{"organizationBuilds": wsLatestBuildsForOrg(orgName)},
	}
}

func (s *webSocket) getLastBuilds(data map[string]interface{}) {
	s.sender <- &Message{
		Mutation: "LAST_BUILDS",
		Data:     map[string]interface{}{"builds": wsLatestBuilds()},
	}
}

func dataID(data map[string]interface{}) (id int64, err error) {
	strID, ok := data["id"].(string)
	if !ok {
		logger.Println("missing id")
		return
	}

	id, err = strconv.ParseInt(strID, 10, 64)
	if err != nil {
		logger.Println("parsing id", err)
	}

	return
}

func (s *webSocket) getBuild(data map[string]interface{}) {
	strID, ok := data["id"].(string)
	if !ok {
		logger.Println("missing id")
		return
	}

	id, err := strconv.Atoi(strID)
	if err != nil {
		logger.Println("parsing id", err)
		return
	}

	projectName, ok := data["projectName"].(string)
	if !ok {
		logger.Println("missing projectName")
		return
	}
	build := wsBuild(projectName, id)
	if build != nil {
		s.sender <- &Message{
			Mutation: "BUILD",
			Data:     map[string]interface{}{"build": build},
		}
	}
}

func (s *webSocket) getBuildLogWatch(data map[string]interface{}) {
	id, err := dataID(data)
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
			s.sender <- &Message{
				Mutation: "BUILD_LOG",
				Data: map[string]interface{}{
					"time": line.Time,
					"line": line.Line,
				}}
		}
	}()
}

func (s *webSocket) getBuildLogUnwatch(data map[string]interface{}) {
	if s.listener != nil {
		logger.Println("Unregister from build log")
		logListenerUnregister <- s.listener
		s.listener = nil
	}
}

// Message encapsulates data sent and received via the websocket.
type Message struct {
	Kind     string
	Action   string `json:"action,omitempty"`
	Mutation string `json:"mutation,omitempty"`
	Data     map[string]interface{}
}

func handleWebSocket(ctx *macaron.Context, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, errorChannel <-chan error) {
	socket := &webSocket{
		receiver:     receiver,
		sender:       sender,
		done:         done,
		disconnect:   disconnect,
		errorChannel: errorChannel,
	}

	socket.mainLoop()
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

func wsBuild(projectName string, buildID int) (build *dbBuild) {
	conn, err := pgxpool.Acquire()
	if err != nil {
		logger.Panic(err)
		return
	}
	defer pgxpool.Release(conn)

	build, err = findBuildByProjectAndID(conn, projectName, buildID)
	if err != nil {
		logger.Println(projectName, buildID, err)
	}

	return
}
