//
// Application go-codetest
//
// Overview:
// A simple go web application for displaying a form and capturing statistics on how the user interacts with it.
// See readme file for details.
//
// Usage:
//		Usage of go-codetest:
//			-p uint
//				port to listen on (default 80)
//
// Build Instructions:
//		1. No external dependencies are required
//		2. Run unit tests
//			 > go test
//		3. Build / Install
//			 > go install
//
// Design Notes:
//
//		The application consists of the following main types:
//			Data 			- stores the user interaction data
//			SessionManager	- maintain a session form (note that a new "session" is created for each load of the form)
//			Server			- main web server
//			client			- client side jQuery page
//
package main

import (
	"flag"
	"log"
	"os"
)

const (
	dftPort = 80 // default listening port
)

func main() {
	//
	// Configuration
	//
	port := flag.Uint("p", dftPort, "port to listen on")
	flag.Parse()
	if flag.NArg() > 0 {
		flag.Usage()
		return
	}

	// configure server then start it listening
	server := &Server{
		Port:       *port,
		sessionMgr: CreateSessionManager(),
		outFile:    os.Stdout,
	}
	log.Fatal(server.Start())
}
