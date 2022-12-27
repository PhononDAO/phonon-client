package orchestrator

import (
	"errors"
)

type PhononTerminal struct {
	sessions []*Session
}

var ErrRemoteNotPaired error = errors.New("no remote card paired")
var ErrNoSession error = errors.New("no connected session with id found")

var globalTerminal *PhononTerminal

func init() {
	globalTerminal = &PhononTerminal{
		sessions: make([]*Session, 0),
	}
}

// NewPhononTerminal returns a new reference to the global phonon terminal singleton.
func NewPhononTerminal() *PhononTerminal {
	return globalTerminal
}

// //
// basic multi-session management
// //

func (t *PhononTerminal) 

func (t *PhononTerminal) ListSessions() []*Session {
	return t.sessions
}

func (t *PhononTerminal) SessionFromID(id string) *Session {
	for _, session := range t.sessions {
		if session.GetCardId() == id {
			return session
		}
	}
	return nil
}

func (t *PhononTerminal) AddSession(sess *Session) {
	for _, session := range t.sessions {
		if session.GetCardId() == sess.GetCardId() {
			return
		}
	}
	t.sessions = append(t.sessions, sess)
}

func (t *PhononTerminal) RemoveSession(sessID string) {
	for index, session := range t.sessions {
		if session.GetCardId() == sessID {
			t.sessions = append(t.sessions[:index], t.sessions[index+1:]...)
			return
		}
	}
}
