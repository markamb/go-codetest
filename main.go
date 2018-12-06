package main

import (
	"log"
	"os"
)

const (
	dftPort = "8080" // default listening port
)

func main() {
	// configure server then start it listening
	server := &Server{
		Port:       dftPort,
		sessionMgr: CreateSessionManager(),
		outFile:    os.Stdout,
	}
	log.Fatal(server.Start())
}
