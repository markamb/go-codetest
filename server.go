package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Data struct {
	WebsiteUrl         string
	SessionId          string
	ResizeFrom         Dimension
	ResizeTo           Dimension
	CopyAndPaste       map[string]bool // map[fieldId]true
	FormCompletionTime int             // Seconds
}

type Dimension struct {
	Width  string
	Height string
}

// PageEvent stores the JSON for API calls
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

func handleMainPage(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		// serve up our single page
		mainTemplate := template.Must(template.ParseFiles("client/index.html"))
		mainTemplate.Execute(response, nil)

	case "POST":
		log.Print("INFO: Form Submitted:\n")

	default:
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// pageHandler processes default page loads.
func pageHandler(response http.ResponseWriter, request *http.Request) {

	// very simple request router used for all requests outside "/api"
	switch request.URL.Path {
	case "":
		fallthrough
	case "/":
		fallthrough
	case "/index.html":
		handleMainPage(response, request)

	default:
		response.WriteHeader(http.StatusNotFound)
		return
	}
}

// apiHandler processes REST API calls
func apiHandler(response http.ResponseWriter, request *http.Request) {

	switch request.Method {
	case "POST":
		event := &PageEvent{}
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(event); err != nil {
			log.Printf("ERROR: Failed to decode request: %v\n", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("API Called: Event = %v\n", *event)
		response.WriteHeader(http.StatusOK)

	default:
		fmt.Printf("Invalid Method Type recieved in API: %v\n\n", request.Method)
		response.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", pageHandler)
	http.HandleFunc("/api", apiHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
