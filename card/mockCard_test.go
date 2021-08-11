package card

import (
	"testing"
)

func TestCardPair(t *testing.T) {
	senderCard, err := NewMockCard()
	if err != nil {
		t.Error(err)
	}

	err = senderCard.InstallCertificate(SignWithDemoKey)
	if err != nil {
		t.Error(err)
	}

	receiverCard, err := NewMockCard()
	if err != nil {
		t.Error(err)
	}
	err = receiverCard.InstallCertificate(SignWithDemoKey)
	if err != nil {
		t.Error(err)
	}

	initPairingData, err := senderCard.InitCardPairing()
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
