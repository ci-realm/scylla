package server

import (
	"runtime"
	"time"

	"github.com/manveru/scylla/queue"
)

var jobQueue *queue.Queue

func SetupQueue() {
	logger.Println("Setting up worker queue")

	jobQueue = &queue.Queue{
		Timeout:    time.Hour,
		CheckEvery: time.Second * 10,
		Pool:       pgxpool,
		Name:       "scylla",
		Retries:    3,
	}

	err := jobQueue.Start(runtime.NumCPU(), func(item *queue.Item) error {
		return runGithubPR(item)
	})
	if err != nil {
		logger.Fatalln(err)
	}
}
