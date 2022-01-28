package orchestrator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/GridPlus/keycard-go/io"
	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/remote"
	"github.com/GridPlus/phonon-client/session"
	"github.com/GridPlus/phonon-client/usb"
	"github.com/GridPlus/phonon-client/util"
	log "github.com/sirupsen/logrus"
)

type PhononTerminal struct {
	sessions []*session.Session
}

type remoteSession struct {
	counterParty model.CounterpartyPhononCard
}

var ErrRemoteNotPaired error = errors.New("no remote card paired")

func (t *PhononTerminal) GenerateMock() error {
	c, err := card.NewMockCard(true, false)
	if err != nil {
		return err
	}
	sess, err := session.NewSession(c)
	if err != nil {
		return err
	}

	t.sessions = append(t.sessions, sess)
	return nil
}

func (t *PhononTerminal) RefreshSessions() ([]*session.Session, error) {
	t.sessions = nil
	var err error
	readers, err := usb.ConnectAllUSBReaders()
	if err != nil {
		return nil, err
	}
	for _, reader := range readers {
		session, err := session.NewSession(card.NewPhononCommandSet(io.NewNormalChannel(reader)))
		if err != nil {
			return nil, err
		}
		t.sessions = append(t.sessions, session)
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

func (t *PhononTerminal) ListSessions() []*session.Session {
	return t.sessions
}

//TODO: probably rework this to move to session because it's awkward to control it from here via the rest server
func (t *PhononTerminal) ConnectRemoteSession(session *session.Session, cardURL string) error {
	u, err := url.Parse(cardURL)
	if err != nil {
		return fmt.Errorf("unable to parse url for card connection: %s", err.Error())
	}
	pathSeparated := strings.Split(u.Path, "/")
	counterpartyID := pathSeparated[len(pathSeparated)-1]
	log.Info("connecting")
	//guard
	remConn, err := remote.Connect(session, fmt.Sprintf("https://%s/phonon", u.Host), true)
	if err != nil {
		return fmt.Errorf("unable to connect to remote session: %s", err.Error())
	}
	log.Info("successfully connected to remote server. Establishing connection to peer")
	err = remConn.ConnectToCard(counterpartyID)
	if err != nil {
		log.Info("returning error from ConnectRemoteSession")
		return err
	}
	localPubKey, err := util.ParseECCPubKey(session.Cert.PubKey)
	if err != nil{
		//we shouldn't get this far and still receive this error
		return err
	}
	if counterpartyID < util.CardIDFromPubKey(localPubKey) {
		paired := make(chan bool, 1)
		go func() {
			for {
				if session.IsPairedToCard() {
					paired <- true
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			return errors.New("pairing timed out")
		case <-paired:
			return nil
		}
	}
	err = session.PairWithRemoteCard(remConn)
	return err
}

func (t *PhononTerminal) SetReceiveMode(sessionIndex int) {
	// set this session to accept incoming secureConnections
}
