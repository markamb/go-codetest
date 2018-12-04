package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

// Server implements our web server logic
type Server struct {
	Port		string
	sessionMgr 	SessionManager
	mainPageTemplate *template.Template
}

// PageEvent stores the JSON from an API call
type PageEvent struct {
	EventType  string `json:"eventType,omitempty"`
	WebsiteUrl string `json:"websiteUrl,omitempty"`
	SessionId  string `json:"sessionId,omitempty"`
	OldWidth   int    `json:"oldWidth,omitempty"`
	OldHeight  int    `json:"oldHeight,omitempty"`
	NewWidth   int    `json:"newWidth,omitempty"`
	NewHeight  int    `json:"newHeight,omitempty"`
	Pasted     bool   `json:"pasted,omitempty"`
	FormId     string `json:"formId,omitempty"`
	Time       int    `json:"time,omitempty"`
}

// processEvent processes an event API call
func processEvent(response http.ResponseWriter, request *http.Request, event *PageEvent, data *Data) {
	switch event.EventType {
	case "resize":
		data.SessionId = event.SessionId
		data.ResizeFrom.Height = event.OldHeight
		data.ResizeFrom.Width = event.OldWidth
		data.ResizeTo.Height = event.NewHeight
		data.ResizeTo.Width = event.NewWidth

	case "copyAndPaste":
		data.CopyAndPaste[event.FormId] = true   // TODO - validate FormId?

	case "timeTaken":
		data.FormCompletionTime = event.Time

	default:
		// this shouldn't happen as events come from our own page
		log.Printf("ERROR: Unexpected EventType: %s", event.EventType)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	data.WebsiteUrl = event.WebsiteUrl
	data.PrintUpdate(event.EventType)		// dump the current data to the screen
}


// handleMainPage processes a request for our 1 (and only) page on the site
func (s *Server) handleMainPage(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		//
		// serve up our single page - note we create a new "session" for every load of the page
		// so the usr interaction data we collect will be reset if the page is refreshed (I'm not clear
		// if this is the desired behaivour?)
		sessionData, err := s.sessionMgr.NewSession()
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.mainPageTemplate.Execute(response, sessionData)

	case "POST":
		//
		// Our page has been submitted.
		// TODO: Get the SessionId from the posted data and display!
//		log.Print("INFO: Form Submitted:\n")

	default:
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// pageHandler processes all page loads.
func (s *Server) pageHandler(response http.ResponseWriter, request *http.Request) {

	// very simple request router used for all requests outside of "/api"
	// We only support a couple paths
	switch request.URL.Path {
	case "":
		fallthrough
	case "/":
		fallthrough
	case "/index.html":
		s.handleMainPage(response, request)

	default:
		response.WriteHeader(http.StatusNotFound)
		return
	}
}

// apiHandler processes REST API calls
func (s *Server) apiHandler(response http.ResponseWriter, request *http.Request) {

	switch request.Method {
	case "POST":
		event := &PageEvent{}
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(event); err != nil {
			log.Printf("ERROR: Failed to decode request: %v\n", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		data, found := s.sessionMgr.Find(event.SessionId)
		if !found {
			// session not found - invalid request or session has expired
			log.Printf("INFO: Invalid or expired session ID recieved: %s\n", event.SessionId)
			response.WriteHeader(http.StatusForbidden)
			return
		}
		processEvent(response, request, event, data)
		response.WriteHeader(http.StatusOK)

	default:
		log.Printf("ERROR: Invalid method type recieved in API: %s\n", request.Method)
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Start setup our routes then starts listening on the required port
func (s *Server) Start() error {
	s.mainPageTemplate = template.Must(template.ParseFiles("client/index.html"))
	http.HandleFunc("/", s.pageHandler)
	http.HandleFunc("/api", s.apiHandler)
	return http.ListenAndServe(":" + DftPort, nil)
}

