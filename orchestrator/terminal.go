package orchestrator

import (
	"errors"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
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
	t.sessions = append(t.sessions, sess)

	return nil
}

func (t *PhononTerminal) RefreshSessions() ([]*card.Session, error) {
	t.sessions = nil
	sessions, err := card.ConnectAll()
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, errors.New("no cards detected")
	}
	//TODO: maybe handle if refresh is called in the middle of a terminal usage
	//Or rename this function to something like InitSessions
	t.sessions = append(t.sessions, sessions...)
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

func (t *PhononTerminal) ConnectRemoteSession(sessionIndex int, someRemoteInterface interface{}) {
	// todo: this whole thing
	// t.sessions[sessionIndex].remote = &remoteSession{}
	return
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
