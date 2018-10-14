package server

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "[scylla] ", log.Lshortfile|log.Ltime|log.Ldate|log.LUTC)
