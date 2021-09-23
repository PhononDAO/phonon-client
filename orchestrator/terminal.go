package orchestrator

import (
	"errors"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
)

type PhononTerminal struct {
	pairings []*Pairing
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
	t.pairings = append(t.pairings, &Pairing{
		s: sess,
	})
	return nil
}

func (t *PhononTerminal) RefreshSessions() error {
	sessions, err := card.ConnectAll()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return errors.New("no cards detected")
	}
	//TODO: maybe handle if refresh is called in the middle of a terminal usage
	//Or rename this function to something like InitSessions
	for _, session := range sessions {
		//TODO: Get friendly name here if possible
		t.pairings = append(t.pairings, &Pairing{s: session})
	}
	return nil
}

func (t *PhononTerminal) InitializePin(sessionIndex int, pin string) error {
	err := t.pairings[sessionIndex].s.Init(pin)
	return err
}

func (t *PhononTerminal) ListSessions() []*card.Session {
	var sessionList []*card.Session
	for _, pairings := range t.pairings {
		sessionList = append(sessionList, pairings.s)
	}
	return sessionList
}

func (t *PhononTerminal) UnlockCard(sessionIndex int, pin string) error {
	// send the pin to the backing card. ezpz
	return nil
}

func (t *PhononTerminal) ListPhonons(cardIndex int) (interface{}, error) {
	// t.pairings[cardIndex].s.ListPhonons()
	return struct{}{}, nil
}

func (t *PhononTerminal) CreatePhonon(cardIndex int) (int, error) {
	// t.pairings[cardIndex].s.cs.CreatePhonon()
	return 0, nil
}

func (t *PhononTerminal) SetDescriptor(cardIndex int, phononIndex int, descriptor interface{}) {
	// todo: replace descriptor with the actual type used for descriptor
	// t.pairings[cardIndex].s.cs.SetDescriptor(phononIndex
}

func (t *PhononTerminal) GetBalance(cardIndex int, phononIndex int) interface{} {
	// It's called GetBalance, but really, it's more of a get filtered phonons from card
	return struct{}{}
}

func (t *PhononTerminal) ConnectRemoteSession(sessionIndex int, someRemoteInterface interface{}) {
	// todo: this whole thing
	t.pairings[sessionIndex].remote = &remoteSession{}
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
	// t.pairings[cardIndex].s.cs.DestroyPhonon()
	return struct{}{}
}
