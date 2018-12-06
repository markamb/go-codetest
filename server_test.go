package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

const (
	testSessionID = "1234ABCD5678" // something we wouldn't normally see
)

type Route int

const (
	DefaultRoute = 0 // page loading handler
	APIRoute     = 1 // API handler
)

// Create default test data to be returned by SessionManager
func dftTestData() *Data {
	return &Data{SessionID: testSessionID, CopyAndPaste: make(map[string]bool)}
}

// serverTestCase defines  the inputs and expected response for a server call
type serverTestCase struct {
	// Destination for output (default to none - disable output)
	outFile *os.File

	// handler to call
	route Route // handler to call (DefaultRoute or APIRoute)

	// request:
	method      string
	URL         string
	requestBody string

	// response expected:
	expectedStatus      int
	expectedContentType string
	testBody            bool     // true if we should validate the response body contents
	bodyMustContain     []string // strings which must be in the response body
	bodyMustNotContain  []string // strings which must NOT be in the response body

	// the following set the expected numbers of calls to methods on the mock SessionManager
	newCalls    int
	findCalls   int
	deleteCalls int

	// test data to return (if null, non is returned, New and Find will fail)
	testData *Data
}

//	<a href="/index.html">See Other</a>

func TestServerGetPage(t *testing.T) {
	// the following request should return the index page with the correct session id inside
	test := &serverTestCase{
		method:              "GET",
		URL:                 "http://localhost/index.html",
		requestBody:         "",
		expectedStatus:      http.StatusOK,
		expectedContentType: "text/html",
		testBody:            true,
		bodyMustContain:     []string{testSessionID, "<html>"},
		bodyMustNotContain:  []string{"{{"},
		newCalls:            1,
		findCalls:           0,
		deleteCalls:         0,
		testData:            dftTestData(),
	}
	testServerRequest(t, test)
}

func TestServerRedirects(t *testing.T) {
	// the following should redirect to main page
	test := &serverTestCase{
		method:              "GET",
		URL:                 "http://localhost",
		requestBody:         "",
		expectedStatus:      http.StatusSeeOther,
		expectedContentType: "text/html",
		testBody:            false,
		bodyMustContain:     nil,
		bodyMustNotContain:  nil,
		newCalls:            0,
		findCalls:           0,
		deleteCalls:         0,
		testData:            dftTestData(),
	}
	testServerRequest(t, test)

	test.URL = "http://localhost/"
	testServerRequest(t, test)
}

func TestServerBadMethod(t *testing.T) {
	test := &serverTestCase{
		method:              "GET",
		URL:                 "http://localhost",
		requestBody:         "",
		testBody:            false,
		expectedStatus:      http.StatusMethodNotAllowed,
		expectedContentType: "",
		newCalls:            0,
		findCalls:           0,
		deleteCalls:         0,
		testData:            dftTestData(),
	}

	test.method = "POST"
	test.URL = "http://localhost"
	testServerRequest(t, test)

	test.URL = "http://localhost/"
	testServerRequest(t, test)

	test.method = "DELETE"
	test.URL = "http://localhost"
	testServerRequest(t, test)

	test.URL = "http://localhost/"
	testServerRequest(t, test)

	test.URL = "http://localhost/index.html"
	testServerRequest(t, test)
}

func TestServerPostForm(t *testing.T) {
	data := url.Values{}
	data.Set("SessionId", testSessionID)
	data.Set("inputEmail", "me@home.com")
	test := &serverTestCase{
		method:              "POST",
		URL:                 "http://localhost/index.html",
		requestBody:         data.Encode(),
		testBody:            false,
		expectedStatus:      http.StatusCreated,
		expectedContentType: "",
		newCalls:            0,
		findCalls:           1,
		deleteCalls:         1,
		testData:            dftTestData(),
	}
	testServerRequest(t, test)
}

func TestServerPostFormBadSessionId(t *testing.T) {

	// first no session id
	data := url.Values{}
	data.Set("BadIDWhichjShouldBeIgnored", "IgnoreMe")
	test := &serverTestCase{
		method:              "POST",
		URL:                 "http://localhost/index.html",
		requestBody:         data.Encode(),
		testBody:            false,
		expectedStatus:      http.StatusForbidden, // Not Authorised!
		expectedContentType: "",
		newCalls:            0,
		findCalls:           1,
		deleteCalls:         0,
		testData:            nil, // Find of session id will fail!
	}
	testServerRequest(t, test)

	// now an invalid one
	data.Set("SessionId", "BADONE")
	test.requestBody = data.Encode()
	testServerRequest(t, test)
}

func TestServerAPIBadEvent(t *testing.T) {
	apiRequest := `{"eventType":"badevent","oldWidth":500,"oldHeight":600,"newWidth":550,` +
		`"newHeight":650,"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}"`
	testAPIRequest(t, apiRequest, http.StatusBadRequest)
}

func TestServerAPIResize(t *testing.T) {
	apiRequest := `{"eventType":"resize","oldWidth":500,"oldHeight":600,"newWidth":550,` +
		`"newHeight":650,"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}"`
	testAPIRequest(t, apiRequest, http.StatusOK)
}

func TestServerAPITimeTaken(t *testing.T) {
	apiRequest := `{"eventType":"timeTaken","time":6,` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` + testSessionID + `"}"`
	testAPIRequest(t, apiRequest, http.StatusOK)
}

func TestServerAPICopyPaste(t *testing.T) {
	apiRequest := `{"eventType":"copyAndPaste","pasted":false,"formId":"inputEmail",` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}`
	testAPIRequest(t, apiRequest, http.StatusOK)

	apiRequest = `{"eventType":"copyAndPaste","pasted":false,"formId":"inputEmail",` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}`
	testAPIRequest(t, apiRequest, http.StatusOK)

	apiRequest = `{"eventType":"copyAndPaste","pasted":false,"formId":"inputCVV",` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}`
	testAPIRequest(t, apiRequest, http.StatusOK)
}

func TestServerAPICopyPasteBadControl(t *testing.T) {
	apiRequest := `{"eventType":"copyAndPaste","pasted":false,"formId":"inputUnknown",` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}`
	testAPIRequest(t, apiRequest, http.StatusBadRequest)
}

//
// Examples to ensure we are writing the correct results to the screen (and accumlating updates)
//

func ExampleServerAPITimeTaken() {
	apiRequest := `{"eventType":"timeTaken","time":6,` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` + testSessionID + `"}"`
	exampleAPIRequest(nil, apiRequest, http.StatusOK)

	//Output:
	//User Data Updated: timeTaken
	//   WebsiteURL: http://localhost:8080/index.html
	//   SessionID: 1234ABCD5678
	//   ResizeFrom: (0,0)
	//   ResizeTo: (0,0)
	//   copyAndPaste controls:
	//   FormCompletionTime: 6 seconds
	//   websiteURLHashCode: 2222077316
}

func ExampleServerAPICopyPaste() {
	apiRequest := `{"eventType":"copyAndPaste","pasted":false,"formId":"inputEmail",` +
		`"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}`
	exampleAPIRequest(nil, apiRequest, http.StatusOK)

	//Output:
	//User Data Updated: copyAndPaste
	//   WebsiteURL: http://localhost:8080/index.html
	//   SessionID: 1234ABCD5678
	//   ResizeFrom: (0,0)
	//   ResizeTo: (0,0)
	//   copyAndPaste controls: inputEmail
	//   FormCompletionTime: 0 seconds
	//   websiteURLHashCode: 2222077316
}

func ExampleServerAPIResize() {
	apiRequest := `{"eventType":"resize","oldWidth":500,"oldHeight":600,"newWidth":550,` +
		`"newHeight":650,"websiteURL":"http://localhost:8080/index.html","sessionID":"` +
		testSessionID + `"}"`
	exampleAPIRequest(nil, apiRequest, http.StatusOK)

	//Output:
	//User Data Updated: resize
	//   WebsiteURL: http://localhost:8080/index.html
	//   SessionID: 1234ABCD5678
	//   ResizeFrom: (500,600)
	//   ResizeTo: (550,650)
	//   copyAndPaste controls:
	//   FormCompletionTime: 0 seconds
	//   websiteURLHashCode: 2222077316
}

//
// Helper functions / types
//

type MockSessionManager struct {

	// used to pass/fail
	t *testing.T

	// delegate all calls to supplied functions
	// If no function supplied and interface is called we fail the test
	newSessionFn func() (*Data, error)
	findFn       func(sessionID string) (*Data, bool)
	deleteFn     func(sessionID string)

	// track number of times each method is called
	newSessionCalls int
	findCalls       int
	deleteCalls     int

	// track parameters for last call made
	findInput   string
	deleteInput string
}

func (s *MockSessionManager) NewSession() (*Data, error) {
	s.newSessionCalls++
	if s.newSessionFn != nil {
		return s.newSessionFn()
	}
	return nil, nil
}

func (s *MockSessionManager) Find(sessionID string) (*Data, bool) {
	s.findCalls++
	s.findInput = sessionID
	if s.findFn != nil {
		return s.findFn(sessionID)
	}
	return nil, false
}

func (s *MockSessionManager) Delete(sessionID string) {
	s.deleteCalls++
	s.deleteInput = sessionID
	if s.deleteFn != nil {
		s.deleteFn(sessionID)
	}
}

// test a POST request to the API and ensure expected response
func testAPIRequest(t *testing.T, requestJSON string, expectedStatus int) {
	test := &serverTestCase{
		route:          APIRoute,
		method:         "POST",
		URL:            "http://localhost/api",
		requestBody:    requestJSON,
		testBody:       false,
		expectedStatus: expectedStatus,
		newCalls:       0,
		findCalls:      1,
		deleteCalls:    0,
		testData:       dftTestData(),
	}
	testServerRequest(t, test)
}

// test a POST request to the API and direct output to stdout
func exampleAPIRequest(t *testing.T, requestJSON string, expectedStatus int) {
	test := &serverTestCase{
		outFile:        os.Stdout,
		route:          APIRoute,
		method:         "POST",
		URL:            "http://localhost/api",
		requestBody:    requestJSON,
		testBody:       false,
		expectedStatus: expectedStatus,
		newCalls:       0,
		findCalls:      1,
		deleteCalls:    0,
		testData:       dftTestData(),
	}
	testServerRequest(t, test)
}

func testServerRequest(t *testing.T, tc *serverTestCase) {

	// disable logging for the duration of the test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stdout)

	// set up test Data object to be returned by session manager
	testData := tc.testData

	// create a mock SessionManager to return our requested testdata, and to
	mockSM := &MockSessionManager{
		t: t,
	}

	if tc.newCalls > 0 {
		mockSM.newSessionFn = func() (*Data, error) {
			if testData == nil {
				return nil, errors.New("not found")
			}
			return testData, nil
		}
	}
	if tc.findCalls > 0 {
		mockSM.findFn = func(sessionId string) (*Data, bool) {
			if testData == nil {
				return nil, false
			}
			return testData, true
		}
	}

	server := &Server{
		outFile:    tc.outFile,
		sessionMgr: mockSM,
	}
	server.Init()

	var req *http.Request
	if len(tc.requestBody) == 0 {
		req = httptest.NewRequest(tc.method, tc.URL, nil)
	} else {
		req = httptest.NewRequest(tc.method, tc.URL, strings.NewReader(tc.requestBody))
	}

	response := httptest.NewRecorder()
	switch tc.route {
	case DefaultRoute:
		server.defaultHandler(response, req)
	case APIRoute:
		server.apiHandler(response, req)
	default:
		t.Fatalf("Invalid Route Requested : %d", tc.route)
	}

	resp := response.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if t == nil {
		return // don't validate results (probably run in an example)
	}

	//
	// check we have the expected response
	//
	if mockSM.newSessionCalls != tc.newCalls {
		t.Errorf("Unexpected numbers of calls to NewSession for request %s %s: expected %d, had %d", tc.method, tc.URL, tc.newCalls, mockSM.newSessionCalls)
	}
	if mockSM.deleteCalls != tc.deleteCalls {
		t.Errorf("Unexpected numbers of calls to Delete for request %s %s: expected %d, had %d", tc.method, tc.URL, tc.deleteCalls, mockSM.deleteCalls)
	}
	if mockSM.findCalls != tc.findCalls {
		t.Errorf("Unexpected numbers of calls to Find for request %s %s: expected %d, had %d", tc.method, tc.URL, tc.findCalls, mockSM.findCalls)
	}
	if resp.StatusCode != tc.expectedStatus {
		t.Errorf("Unexpected status code for request %s %s: expected %d, got %d", tc.method, tc.URL, tc.expectedStatus, resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); len(tc.expectedContentType) != 0 && !strings.Contains(ct, tc.expectedContentType) {
		t.Errorf("Unexpected Content-Type for request %s %s: expected %s, got %s", tc.method, tc.URL, tc.expectedContentType, ct)
	}
	if tc.testBody {
		// test the body contents - we just look for presence and absence of certain strings
		bStr := string(body)
		for _, next := range tc.bodyMustContain {
			if !strings.Contains(bStr, next) {
				t.Errorf("failed to find expected text in response body for request %s %s: expected to find %s", tc.method, tc.URL, next)
			}
		}
		for _, next := range tc.bodyMustNotContain {
			if strings.Contains(bStr, next) {
				t.Errorf("found uexpected text in response body for request %s %s: found %s", tc.method, tc.URL, next)
			}
		}
	}
}
