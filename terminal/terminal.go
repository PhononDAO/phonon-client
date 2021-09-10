package terminal

import "github.com/GridPlus/phonon-client/card"

type terminal struct {
	plist []pairing
}
type pairing struct {
	s *card.Session
	r *remoteSession
}

type remoteSession struct {

}

func (t *terminal) RefreshSessions(){
	// list all cards
	// start a session for each one of them
}

func (t *terminal) InitializePin(sessionIndex int, pin string) error {
	err := t.plist[sessionIndex].s.Init(pin)
	return err

}

func (t *terminal) ListSessions()[]*card.Session {
	var sessionList []*card.Session
	for _, pairings := range(t.plist){
		sessionList = append(sessionList,pairings.s)
	}
	return sessionList
}

func (t *terminal) UnlockCard(sessionIndex int, pin string) {
	// send the pin to the backing card. ezpz
}

func (t *terminal) ListPhonons(cardIndex int) {
	// t.plist[cardIndex].s.ListPhonons()
}

func (t *terminal) CreatePhonon(cardIndex int) {
	// t.plist[cardIndex].s.cs.CreatePhonon()
}

func (t *terminal) Setdescriptor(cardIndex int, phononIndex int, descriptor interface{}) {
	// todo: replace descriptor with the actual type used for descriptor
	// t.plist[cardIndex].s.cs.SetDescriptor(phononIndex
}

func (t *terminal) GetBalance(cardIndex int, phononIndex  int) {
	// It's called GetBalance, but really, it's more of a get filtered phonons from card

}

func (t *terminal) ConnectRemoteSession(sessionIndex int, someRemoteInterface interface{}) {
	// todo: this whole thing
		t.plist[sessionIndex].r = &remoteSession{}
}

func (t *terminal) ProposeTransaction() {
	// implementation details to be determined at a later date
}

func (t *terminal) ListReceivedProposedTransactions() {
	// implementation details to be determined at a later date
}

func (t *terminal) SendPhonons(sessionIndex int, phononIndexes []int) {
	// check for remote session
	// start or check for secure card-card connection
	// send the phonons
}

func (t *terminal) SetReceivemode(sessionIndex int) {
	// set this session to accept incoming secureConnections
}
/* not sure how we should handle invoice requests
func (t *termianl) ApproveInvoice() {
	//todo
}*/

func (t *terminal) RedeemPhonon(cardIndex int, phononIndex int) {
	// t.plist[i].s.cs.DestroyPhonon()
}

func (t *terminal) GenerateMock()error {
	c, err := card.NewMockCard()
	if err != nil{
		return err
	}
	sess := card.NewSession(c,true)
	t.plist = append(t.plist,pairing{
		s: sess,
	})
	return nil
}
