package main

import (
	"log"
)

const (
	dftPort = "8080" // default listening port
)

func main() {
	// configure server then start it listening
	server := &Server{
		Port:       dftPort,
		sessionMgr: CreateSessionManager(),
	}
	log.Fatal(server.Start())
}
