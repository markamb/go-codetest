package main

import (
	"log"
)

const (
	DftPort = "8080"		// default listening port
)

func main() {
	// configure server then start it listening
	server := &Server {
		Port: DftPort,
		sessionMgr: CreateSessionManager(),
	}
	log.Fatal(server.Start())
}
