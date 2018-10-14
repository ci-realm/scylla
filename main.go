package main

import (
	"log"
	"os"

	"github.com/manveru/scylla/server"
)

var logger = log.New(os.Stderr, "[main] ", log.Lshortfile|log.Ltime|log.Ldate|log.LUTC)

func init() {
	err := os.MkdirAll(os.TempDir(), os.FileMode(0755))
	if err != nil {
		logger.Fatalln("failed making ", os.TempDir(), err)
	}
}

func main() {
	server.Start()
}
