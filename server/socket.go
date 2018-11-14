package server

import (
	"fmt"
	"strconv"

	"github.com/manveru/scylla/queue"
	macaron "gopkg.in/macaron.v1"
)

func handleWebSocket(ctx *macaron.Context, receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, errorChannel <-chan error) {
	socket := &webSocket{
		ctx:          ctx,
		receiver:     receiver,
		sender:       sender,
		done:         done,
		disconnect:   disconnect,
		errorChannel: errorChannel,
	}

	socket.mainLoop()
}

// Message encapsulates data sent and received via the websocket.
type Message struct {
	Kind     string `json:"kind,omitempty"`
	Mutation string `json:"mutation,omitempty"`
	Data     data   `json:"data"`
}

type data map[string]interface{}

func (d data) getString(key string) (string, error) {
	value, ok := d[key].(string)
	if ok {
		return value, nil
	}
	return value, fmt.Errorf("Coudln't find key '%s' in Data %v", key, d)
}

func (d data) getInt64(key string) (int64, error) {
	found, ok := d[key]
	if !ok {
		return 0, fmt.Errorf("Couldn't find key '%s' in Data %v", key, d)
	}

	value, ok := found.(string)
	if !ok {
		return 0, fmt.Errorf("Couldn't transform value of key %s: '%#v' into int", key, found)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)

	return parsed, err
}

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
		case <-s.done:
			return
		case err := <-s.errorChannel:
			logger.Println(err)
		}
	}
}

func (s *webSocket) getOrganizations(data map[string]interface{}) {
	s.sender <- &Message{
		Mutation: "organizations",
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
		Mutation: "organizationBuilds",
		Data:     map[string]interface{}{"organizationBuilds": wsLatestBuildsForOrg(orgName)},
	}
}

func (s *webSocket) restartBuild(d data) {
	buildID, err := d.getInt64("id")
	if err != nil {
		logger.Println(err)
	}
	item := &queue.Item{Args: map[string]interface{}{"build_id": buildID, "Host": progressHost(s.ctx)}}
	err = jobQueue.Insert(item)
	if err != nil {
		logger.Println(err)
	}
}

func (s *webSocket) getLastBuilds(_ data) {
	s.sender <- &Message{
		Mutation: "lastBuilds",
		Data:     map[string]interface{}{"builds": wsLatestBuilds()},
	}
}

func dataID(data data) (id int64, err error) {
	strID, err := data.getString("id")
	if err != nil {
		logger.Println(err)
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
			Mutation: "build",
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
				Mutation: "buildLog",
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
