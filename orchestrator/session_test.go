package orchestrator

import (
	"testing"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
)

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
