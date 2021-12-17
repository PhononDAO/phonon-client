package orchestrator

import (
	"testing"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/session"
)

func TestCardToCardPair(t *testing.T) {
	//Test with real sender and mock receiver card
	cs, err := card.Connect(0)
	if err != nil {
		t.Error(err)
		return
	}
	s, err := session.NewSession(cs)
	if err != nil {
		t.Error(err)
		return
	}
	mockCard, err := card.NewMockCard(true, false)
	if err != nil {
		t.Error(err)
		return
	}

	mockSession, err := session.NewSession(mockCard)
	if err != nil {
		t.Error(err)
		return
	}
	err = mockSession.VerifyPIN("111111")
	if err != nil {
		t.Error(err)
		return
	}
	mockRemote := NewLocalCounterParty(mockSession)

	err = s.VerifyPIN("111111")
	if err != nil {
		t.Error(err)
		return
	}
	err = s.PairWithRemoteCard(mockRemote)
	if err != nil {
		t.Error("error pairing with remote: ", err)
		return
	}
	t.Log("paired local actual card with remote mock")

	//Test with real receiver and mock sender card

	cardAsRemote := NewLocalCounterParty(s)
	err = mockSession.PairWithRemoteCard(cardAsRemote)
	if err != nil {
		t.Error("error pairing mock with remote card: ", err)
		return
	}
	t.Log("paired local mock with remote actual card")
}
