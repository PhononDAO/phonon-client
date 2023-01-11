package orchestrator_test

import (
	"testing"

	"github.com/GridPlus/phonon-core/pkg/backend/mock"
	"github.com/GridPlus/phonon-core/pkg/model"
	"github.com/GridPlus/phonon-core/pkg/orchestrator"
	log "github.com/sirupsen/logrus"
)

func TestE2ELocalSendPhonon(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mock2, _ := mock.NewMockCard(true, false)
	mock1, _ := mock.NewMockCard(true, false)

	sess1, _ := orchestrator.NewSession(mock1)
	sess2, _ := orchestrator.NewSession(mock2)

	sess1.VerifyPIN("111111")
	sess2.VerifyPIN("111111")

	sess1.ConnectToLocalProvider()
	sess2.ConnectToLocalProvider()

	sess1.ConnectToCounterparty(sess2.GetCardId())
	sess2.ConnectToCounterparty(sess1.GetCardId())

	keyIndex, _, err := sess2.CreatePhonon()
	if err != nil {
		log.Fatal("unable to create phonon: ", err)
	}
	sess2.SetDescriptor(&model.Phonon{
		KeyIndex:  keyIndex,
		CurveType: 0,
		Denomination: model.Denomination{
			Base:     1,
			Exponent: 3,
		},
		CurrencyType: 2,
	})
	err = sess2.SendPhonons([]model.PhononKeyIndex{keyIndex})
	if err != nil {
		t.Error("session 2 could not send initial phonon: ", err)
	}
	phonons, err := sess1.ListPhonons(0, 0, 0)
	if err != nil {
		t.Error("could not list phonons on session 1: ", err)
	}
	t.Logf("phonons: %+v", phonons)
	err = sess1.SendPhonons([]model.PhononKeyIndex{0})
	if err != nil {
		t.Error("session 1 could not return received phonon: ", err.Error())
	}

}
