package orchestrator_test

import (
	"testing"

	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/sirupsen/logrus"
)

func TestCardToCardPair(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	term := orchestrator.NewPhononTerminal()
	sessions, err := term.RefreshSessions()
	if err != nil {
		t.Error(err)
		return
	}
	s := sessions[0]
	mockID, err := term.GenerateMock()
	if err != nil {
		t.Error(err)
		return
	}

	mockSession := term.SessionFromID(mockID)
	if mockSession == nil {
		t.Error("unable to locate newly generated mock session")
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
	err = s.ConnectToLocalProvider()
	if err != nil {
		t.Error(err)
		return
	}
	err = mockSession.ConnectToLocalProvider()
	if err != nil {
		t.Error(err)
		return
	}
	err = s.ConnectToCounterparty(mockID)
	if err != nil {
		t.Error(err)
		return
	}

}
