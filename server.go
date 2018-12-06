package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
)

const (
	sessionIDControl = "sessionID"
	mainPageURL      = "/index.html"
	apiURL           = "/api"
)

// formControls contains a set of all valid form control ids
var validControls = map[string]bool{
	"inputEmail":      true,
	"inputCVV":        true,
	"inputCardNumber": true,
}

// Server implements our web server logic
type Server struct {
	Port             string
	sessionMgr       SessionManager
	outFile          *os.File // file to send out put to (default to stdout)
	mainPageTemplate *template.Template
}

// PageEvent stores the JSON from an API call
type PageEvent struct {
	EventType  string `json:"eventType,omitempty"`
	WebsiteURL string `json:"websiteUrl,omitempty"`
	SessionID  string `json:"sessionId,omitempty"`
	OldWidth   int    `json:"oldWidth,omitempty"`
	OldHeight  int    `json:"oldHeight,omitempty"`
	NewWidth   int    `json:"newWidth,omitempty"`
	NewHeight  int    `json:"newHeight,omitempty"`
	Pasted     bool   `json:"pasted,omitempty"`
	FormID     string `json:"formId,omitempty"`
	Time       int    `json:"time,omitempty"`
}

// processEvent processes an event API call
func (s *Server) processEvent(response http.ResponseWriter, request *http.Request, event *PageEvent, data *Data) {
	data.mutex.Lock()
	defer data.mutex.Unlock()

	switch event.EventType {
	case "resize":
		data.SessionID = event.SessionID
		data.ResizeFrom.Height = event.OldHeight
		data.ResizeFrom.Width = event.OldWidth
		data.ResizeTo.Height = event.NewHeight
		data.ResizeTo.Width = event.NewWidth

	case "copyAndPaste":
		if _, found := validControls[event.FormID]; !found {
			log.Printf("ERROR: Unexpected form ID: %s", event.FormID)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		data.CopyAndPaste[event.FormID] = true

	case "timeTaken":
		data.FormCompletionTime = event.Time

	default:
		// this shouldn't happen as events come from our own page
		log.Printf("ERROR: Unexpected EventType: %s", event.EventType)
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	data.WebsiteURL = event.WebsiteURL
	data.PrintUpdate(s.outFile, event.EventType) // dump the current data to the screen
	response.WriteHeader(http.StatusOK)
}

// processMainPageGet processes a GET on our main page
// serve up our single page - note we create a new "session" for every load of the page
// so the user interaction data we collect will be reset if the page is refreshed.
func (s *Server) processMainPageGet(response http.ResponseWriter, request *http.Request) {
	sessionData, err := s.sessionMgr.NewSession()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.mainPageTemplate.Execute(response, sessionData)
}

// processMainPagePost processes a POST on our main page
func (s *Server) processMainPagePost(response http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	sid := request.FormValue(sessionIDControl)
	data, found := s.sessionMgr.Find(sid)
	if !found {
		// session not found - invalid request or session has expired
		log.Printf("INFO: Invalid or expired session ID recieved: %s\n", sid)
		response.WriteHeader(http.StatusForbidden)
		return
	}
	data.PrintUpdate(s.outFile, "(Form Posted)")
	s.sessionMgr.Delete(sid) // delete this session once form is submitted
	response.WriteHeader(http.StatusCreated)
}

// processMainPage processes a request for our 1 (and only) page on the site
func (s *Server) processMainPage(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		s.processMainPageGet(response, request)
	case "POST":
		s.processMainPagePost(response, request)
	default:
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// defaultPageHandler processes all page loads.
func (s *Server) defaultHandler(response http.ResponseWriter, request *http.Request) {
	// very simple request router used for all requests outside of api and index.html
	switch request.URL.Path {
	case "":
		fallthrough
	case "/":
		if request.Method == "GET" {
			http.Redirect(response, request, mainPageURL, http.StatusSeeOther)
		} else {
			response.WriteHeader(http.StatusMethodNotAllowed)
		}
	case mainPageURL:
		s.processMainPage(response, request)
	default:
		response.WriteHeader(http.StatusNotFound)
		return
	}
}

// apiHandler processes API calls
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
		data, found := s.sessionMgr.Find(event.SessionID)
		if !found {
			// session not found - invalid request or session has expired
			log.Printf("INFO: Invalid or expired session ID recieved: %s\n", event.SessionID)
			response.WriteHeader(http.StatusForbidden)
			return
		}
		s.processEvent(response, request, event, data)

	default:
		log.Printf("ERROR: Invalid method type recieved in API: %s\n", request.Method)
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Init initialises the server ready for use.
func (s *Server) Init() {
	s.mainPageTemplate = template.Must(template.ParseFiles("client/index.html"))
}

// Start setup our routes then starts listening on the required port
func (s *Server) Start() error {
	s.Init()
	http.HandleFunc(apiURL, s.apiHandler)
	//	http.HandleFunc(mainPageURL, s.mainPageHandler)
	http.HandleFunc("/", s.defaultHandler)
	return http.ListenAndServe(":"+s.Port, nil)
}
