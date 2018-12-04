package main

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

// SessionManager is an interface to maintains a map of session ids to Session types
// Note that a session is used to store types of type "Data"
type SessionManager interface {
	NewSession() (*Data, error)
	Find(sessionId string) (*Data, bool)
	Delete(sessionId string)
}

// CreateSessionManager returns a new SessionManger implementation
func CreateSessionManager() SessionManager {
	return &DataSessionManager{
		sessions: make(map[string]*Session),
	}
}

// Session stores all the per session data we require
type Session struct {
	data   	*Data		// our tracked data for this session
	//	time.Time	expiry		// TODO: monitor when session expires then delete in background thread
}

// DataSessionManager is a thread safe type implementing the SessionManger interface
type DataSessionManager struct {
	sessions	map[string]*Session
	//	maxLifetime	int64		// TODO - expiry of stale sessions
	mutex		sync.Mutex
}

// NewSession creates a new session with a random session id and adds it to
// this session manager. Returns the new Data on success, or an error on failure
func (m *DataSessionManager) NewSession() (*Data, error) {
	id, err := makeSessionId()
	if err != nil {
		return nil, err
	}
	d := &Session {
		&Data{SessionId: id,
			CopyAndPaste: make(map[string]bool),
		},
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.sessions[id] = d
	return d.data, nil
}

// Find returns the Data stored for the given SessionId or nil if none exists
// Returns the session if found, and a flag to indicate success
func (m *DataSessionManager) Find(sessionId string) (*Data, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s, ok := m.sessions[sessionId]
	if s == nil || !ok {
		return nil, false
	}
	return s.data, true
}

// Delete removes the specified session id if present (no effect if not found)
func (m *DataSessionManager) Delete(sessionId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.sessions, sessionId)
}

// makeSessionId generate a new random session id
func makeSessionId() (string, error) {
	key := make([]byte, 64)
	// TODO - do we need to set a seded? Could do this in SessionMager construction
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(key), nil
}
