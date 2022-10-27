package orchestrator_test

import (
	"testing"

	// "github.com/GridPlus/phonon-client/card"
	// "github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/config"
	"github.com/GridPlus/phonon-client/orchestrator"

	// "github.com/GridPlus/phonon-client/util"
	"github.com/sirupsen/logrus"
)

func TestCardToCardPair(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	term := orchestrator.NewPhononTerminal(config.DefaultConfig())
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

// //Integration tests that the card actually validates the certificate of it's counterparty during pairing.
// func TestCardValidatesCounterpartyCert(t *testing.T) {
// 	m, _ := card.NewMockCard(false, false)

// 	signingKey, err := util.ParseECCPrivKey(cert.PhononMockCAPrivKey)
// 	if err != nil {
// 		t.Error("error parsing private key. err: ", err)
// 	}
// 	testSigner := cert.GetSignerWithPrivateKey(*signingKey)
// 	_, _, _, err = m.Select()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = m.InstallCertificate(testSigner)
// 	if err != nil {
// 		t.Fatal("unable to install mock cert")
// 	}

// 	err = m.Init("111111")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sender, err := orchestrator.NewSession(m)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = sender.VerifyPIN("111111")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	receiverCard, err := card.Connect(0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	receiver, err := orchestrator.NewSession(receiverCard)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = receiver.VerifyPIN("111111")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	sender.ConnectToLocalProvider()
// 	receiver.ConnectToLocalProvider()

// 	err = sender.ConnectToCounterparty(receiver.GetName())
// 	if err == nil {
// 		t.Fatal("pairing mock sender with bad cert with real receiver should have failed but didn't")
// 	}

// 	//Try it the other way around, restarting sessions
// 	mockReceiver, err := orchestrator.NewSession(m)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = mockReceiver.VerifyPIN("111111")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	senderCard, err := card.Connect(0)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	sender, err = orchestrator.NewSession(senderCard)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = sender.VerifyPIN("111111")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// counterParty = NewLocalCounterParty(mockReceiver)

// 	sender.ConnectToLocalProvider()
// 	receiver.ConnectToLocalProvider()

// 	err = sender.ConnectToCounterparty(receiver.GetName())
// 	if err == nil {
// 		t.Fatal("pairing real sender with mock receiver with bad cert should have failed but didn't")
// 	}

// 	t.Log("pairing real sender with mock receiver resulted in correct err: ", err)
// 	//Real cards should not be able to validate this mock in integration tests because they will be installed with the demo or alpha cert

// 	// privKey, _ := ethcrypto.GenerateKey()
// 	// fmt.Println("private key: ", util.ECCPrivKeyToHex(privKey))
// 	// fmt.Println("public key: ", util.ECCPubKeyToHexString(&privKey.PublicKey))
// }
