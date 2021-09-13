package orchestrator

import (
	"errors"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
)

type PhononTerminal struct {
	pairings []*Pairing
}

type Pairing struct {
	s      *card.Session
	remote *remoteSession
	name   string
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
	sess := card.NewSession(c, true)
	t.pairings = append(t.pairings, &Pairing{
		s: sess,
	})
	return nil
}

func (t *PhononTerminal) RefreshSessions() {
	// list all cards
	// start a session for each one of them
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

func (t *PhononTerminal) UnlockCard(sessionIndex int, pin string) {
	// send the pin to the backing card. ezpz
}

func (t *PhononTerminal) ListPhonons(cardIndex int) {
	// t.pairings[cardIndex].s.ListPhonons()
}

func (t *PhononTerminal) CreatePhonon(cardIndex int) {
	// t.pairings[cardIndex].s.cs.CreatePhonon()
}

func (t *PhononTerminal) Setdescriptor(cardIndex int, phononIndex int, descriptor interface{}) {
	// todo: replace descriptor with the actual type used for descriptor
	// t.pairings[cardIndex].s.cs.SetDescriptor(phononIndex
}

func (t *PhononTerminal) GetBalance(cardIndex int, phononIndex int) {
	// It's called GetBalance, but really, it's more of a get filtered phonons from card

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

func (t *PhononTerminal) SetReceivemode(sessionIndex int) {
	// set this session to accept incoming secureConnections
}

/* not sure how we should handle invoice requests
func (t *termianl) ApproveInvoice() {
	//todo
}*/

func (t *PhononTerminal) RedeemPhonon(cardIndex int, phononIndex int) {
	// t.pairings[i].s.cs.DestroyPhonon()
}