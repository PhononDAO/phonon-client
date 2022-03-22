package orchestrator

import (
	"testing"

	"github.com/GridPlus/phonon-client/card"
)

func TestCardToCardPair(t *testing.T) {
	//Test with real sender and mock receiver card
	cs, err := card.Connect(0)
	if err != nil {
		t.Error(err)
		return
	}
	s, err := NewSession(cs)
	if err != nil {
		t.Error(err)
		return
	}
	mockCard, err := card.NewMockCard(true, false)
	if err != nil {
		t.Error(err)
		return
	}

	mockSession, err := NewSession(mockCard)
	if err != nil {
		t.Error(err)
		return
	}
	err = mockSession.VerifyPIN("111111")
	if err != nil {
		t.Error(err)
		return
	}
	err = s.VerifyPIN("111111")
	if err != nil {
		t.Error(err)
		return
	}
	s.ConnectToLocalProvider()
	mockSession.ConnectToLocalProvider()

	s.ConnectToCounterparty(mockSession.GetName())

}
