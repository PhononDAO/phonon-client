package orchestrator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/remote"
	log "github.com/sirupsen/logrus"
)

type PhononTerminal struct {
	sessions []*card.Session
}

type remoteSession struct {
	counterParty model.CounterpartyPhononCard
}

var ErrRemoteNotPaired error = errors.New("no remote card paired")

func (t *PhononTerminal) GenerateMock() error {
	c, err := card.NewMockCard()
	if err != nil {
		return err
	}
	sess, _ := card.NewSession(c)
	// sign with demo key. there's no reason a mock card would not be signed with demo key
	err = c.InstallCertificate(cert.SignWithDemoKey)
	if err != nil {
		return err
	}
	err = sess.Init("111111")
	if err != nil {
		return err
	}
	t.sessions = append(t.sessions, sess)
	return nil
}

func (t *PhononTerminal) RefreshSessions() ([]*card.Session, error) {
	t.sessions = nil
	var err error
	t.sessions, err = card.ConnectAll()
	if err != nil {
		return nil, err
	}
	if len(t.sessions) == 0 {
		return nil, errors.New("no cards detected")
	}
	return t.sessions, nil
}

// func (t *PhononTerminal) InitializePin(sessionIndex int, pin string) error {
// 	err := t.sessions[sessionIndex].Init(pin)
// 	return err
// }

func (t *PhononTerminal) ListSessions() []*card.Session {
	return t.sessions
}

func (t *PhononTerminal) ConnectRemoteSession(session *card.Session, cardURL string) error {
	u, err := url.Parse(cardURL)
	if err != nil {
		return fmt.Errorf("Unable to parse url for card connection: %s", err.Error())
	}
	pathSeparated := strings.Split(u.Path, "/")
	counterpartyID := pathSeparated[len(pathSeparated)-1]
	log.Info("connecting")
	remConn, err := remote.Connect(session, fmt.Sprintf("https://%s/phonon", u.Host), true)
	if err != nil {
		return fmt.Errorf("Unable to connect to remote session: %s", err.Error())
	}
	log.Info("successfully connected to remote server. Establishing connection to peer")
	err = remConn.ConnectToCard(counterpartyID)
	if err != nil {
		return err
	}
	if counterpartyID < session.GetName() {
		return nil
	}
	err = session.PairWithRemoteCard(remConn)
	return err
}

func (t *PhononTerminal) SetReceiveMode(sessionIndex int) {
	// set this session to accept incoming secureConnections
}
