package orchestrator

import (
	"testing"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/remote/v1/server"
)

/*
func TestDepositPhonons(t *testing.T) {
	mock, _ := card.NewMockCard(true, false)
	s, _ := NewSession(mock)
	s.VerifyPIN("111111")

	//Test Single Ethereum Deposit
	denom, _ := model.NewDenomination(1)
	phonons, err := s.InitDepositPhonons(model.Ethereum, []model.Denomination{denom})
	if err != nil {
		t.Error("failed to initiate phonon deposit. err: ", err)
	}
	t.Log(phonons)
}
*/
func TestE2EJumpboxSendPhonon(t *testing.T) {
	//todo: fix this
	server.StartServer("42069", "/Users/nate/Documents/localhost.cer.pem", "/Users/nate/Documents/localhost.key.pem")
	term := NewPhononTerminal()
	mock1, _ := term.GenerateMock()
	mock2, _ := term.GenerateMock()
	sess1 := term.SessionFromID(mock1)
	sess2 := term.SessionFromID(mock2)
	sess1.VerifyPIN("111111")
	sess2.VerifyPIN("111111")
	sess1.ConnectToRemoteProvider("https://localhost:42069/phonon")
	sess2.ConnectToRemoteProvider("https://localhost:42069/phonon")
	sess1.ConnectToCounterparty(mock2)
	sess1.CreatePhonon()
	sess1.SetDescriptor(&model.Phonon{
		KeyIndex:  0,
		CurveType: 0,
		Denomination: model.Denomination{
			Base:     1,
			Exponent: 3,
		},
		CurrencyType: 2,
	})
	sess1.SendPhonons([]uint16{
		0,
	})
}
