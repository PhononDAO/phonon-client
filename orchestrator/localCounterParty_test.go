package orchestrator

import (
	"testing"

	"github.com/GridPlus/phonon-client/card"
)

func TestCardToCardPair(t *testing.T) {
	//Test with real sender and mock receiver card
	cs, err := Connect()
	session, err := card.NewSession(cs)
	if err != nil {
		t.Error(err)
		return
	}
	mockCard, err := card.NewInitializedMockCard()
	if err != nil {
		t.Error(err)
		return
	}

	mockSession, err := card.NewSession(mockCard)
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

	err = session.VerifyPIN("111111")
	if err != nil {
		t.Error(err)
		return
	}
	err = session.PairWithRemoteCard(mockRemote)
	if err != nil {
		t.Error("error pairing with remote: ", err)
		return
	}
	t.Log("paired local actual card with remote mock")

	//Test with real receiver and mock sender card

	cardAsRemote := NewLocalCounterParty(session)
	err = mockSession.PairWithRemoteCard(cardAsRemote)
	if err != nil {
		t.Error("error pairing mock with remote card: ", err)
		return
	}
	t.Log("paired local mock with remote actual card")
}
