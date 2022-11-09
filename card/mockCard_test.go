package card

import (
	"github.com/GridPlus/phonon-client/cert"
	"testing"
)

func TestCardPair(t *testing.T) {
	senderCard, err := NewMockCard(false, false)
	if err != nil {
		t.Error(err)
	}

	err = senderCard.InstallCertificate(cert.SignWithDemoKey)
	if err != nil {
		t.Error(err)
	}

	receiverCard, err := NewMockCard(false, false)
	if err != nil {
		t.Error(err)
	}
	err = receiverCard.InstallCertificate(cert.SignWithDemoKey)
	if err != nil {
		t.Error(err)
	}

	initPairingData, err := senderCard.InitCardPairing(receiverCard.IdentityCert)
	if err != nil {
		t.Error("error in initCardPairing")
		t.Error(err)
	}
	_, err = receiverCard.CardPair(initPairingData)
	if err != nil {
		t.Error("error in card pair")
		t.Error(err)
	}
}
