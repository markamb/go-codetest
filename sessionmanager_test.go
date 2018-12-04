package main

import "testing"

func TestSessionMangerAddFindDelete(t *testing.T) {

	sm := CreateSessionManager()

	// create a couple new sessions
	s1, err := sm.NewSession()
	if err != nil || s1 == nil || len(s1.SessionId) == 0 {
		t.Errorf("SessionManger: Failed to create new session! (%v, %v)", err, s1)
	}
	s2, err := sm.NewSession()
	if err != nil || s2 == nil || len(s2.SessionId) == 0 {
		t.Errorf("SessionManger: Failed to create new session! (%v, %v)", err, s1)
	}

	// load and check correct
	s, found := sm.Find(s1.SessionId)
	if !found {
		t.Errorf("SessionManger: Failed to find session! (%s)", s1.SessionId)
	}
	if s != s1 {
		t.Errorf("SessionManger: Failed to load same session! (%p, %p)", s1, s)
	}
	s, found = sm.Find(s2.SessionId)
	if !found {
		t.Errorf("SessionManger: Failed to find session! (%s)", s2.SessionId)
	}
	if s != s2 {
		t.Errorf("SessionManger: Failed to load same session! (%p, %p)", s2, s)
	}

	// delete, load then check
	sm.Delete(s1.SessionId)
	s, found = sm.Find(s1.SessionId)
	if found || s != nil {
		t.Errorf("SessionManger: Found deleted session! (%s, %p)", s1.SessionId, s)
	}
	s, found = sm.Find(s2.SessionId)
	if !found {
		t.Errorf("SessionManger: Failed to find session! (%s)", s2.SessionId)
	}
	if s != s2 {
		t.Errorf("SessionManger: Failed to load same session! (%p, %p)", s2, s)
	}

	// delete last one
	sm.Delete(s2.SessionId)
	s, found = sm.Find(s2.SessionId)
	if found || s != nil {
		t.Errorf("SessionManger: Found deleted session! (%s, %p)", s2.SessionId, s)
	}
}
