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
	if err != nil{
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
	//TODO: maybe handle if refresh is called in the middle of a terminal usage
	//Or rename this function to something like InitSessions
	return t.sessions, nil
}

// func (t *PhononTerminal) InitializePin(sessionIndex int, pin string) error {
// 	err := t.sessions[sessionIndex].Init(pin)
// 	return err
// }

func (t *PhononTerminal) ListSessions() []*card.Session {
	return t.sessions
}

func (t *PhononTerminal) UnlockCard(sessionIndex int, pin string) error {
	// send the pin to the backing card. ezpz
	return nil
}

func (t *PhononTerminal) ListPhonons(cardIndex int) (interface{}, error) {
	// t.sessions[cardIndex].s.ListPhonons()
	return struct{}{}, nil
}

func (t *PhononTerminal) CreatePhonon(cardIndex int) (int, error) {
	// t.sessions[cardIndex].s.cs.CreatePhonon()
	return 0, nil
}

func (t *PhononTerminal) SetDescriptor(cardIndex int, phononIndex int, descriptor interface{}) {
	// todo: replace descriptor with the actual type used for descriptor
	// t.sessions[cardIndex].s.cs.SetDescriptor(phononIndex
}

func (t *PhononTerminal) GetBalance(cardIndex int, phononIndex int) interface{} {
	// It's called GetBalance, but really, it's more of a get filtered phonons from card
	return struct{}{}
}

func (t *PhononTerminal) ConnectRemoteSession(session *card.Session,cardURL string) error {
	u, err := url.Parse(cardURL)
	if err != nil{
		return fmt.Errorf("Unable to parse url for card connection: %s", err.Error())
	}
	pathSeparated := strings.Split(u.Path,"/")
	counterpartyID := pathSeparated[len(pathSeparated)-1]
	log.Info("connecting")
	remConn, err := remote.Connect(session, fmt.Sprintf("https://%s/phonon",u.Host), true)
	if err != nil {
		return fmt.Errorf("Unable to connect to remote session: %s", err.Error())
	}
	log.Info("successfully connected to remote server. Establishing connection to peer")
	err = remConn.ConnectToCard(counterpartyID)
	if err != nil{
		return err
	}
	if counterpartyID < session.GetName(){
		return nil
	}
	err = session.PairWithRemoteCard(remConn)
	return err
}

func (t *PhononTerminal) ProposeTransaction() {
	// implementation details to be determined at a later date
}

func (t *PhononTerminal) ListReceivedProposedTransactions() {
	// implementation details to be determined at a later date
}

func (t *PhononTerminal) SetReceiveMode(sessionIndex int) {
	// set this session to accept incoming secureConnections
}

/* not sure how we should handle invoice requests
func (t *termianl) ApproveInvoice() {
	//todo
}*/

func (t *PhononTerminal) RedeemPhonon(cardIndex int, phononIndex int) interface{} {
	// t.sessions[cardIndex].s.cs.DestroyPhonon()
	return struct{}{}
}
