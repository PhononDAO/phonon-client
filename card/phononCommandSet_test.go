package card

import (
	"os"
	"runtime"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
	os.Exit(m.Run())
}

// //Open a secure connection and make sure a pin is set on card
// func TestMain(m *testing.M) {
// 	cs, err := OpenSecureConnection()
// 	testPin := "111111"
// 	err = cs.VerifyPIN(testPin)
// 	if err != nil {
// 		fmt.Println("unable to verify pin: ", err)
// 		return
// 	}
// 	m.Run()
// }

// //Integration tests to check card functionality against
// func TestOpenSecureConnection(t *testing.T) {
// 	cs, err := OpenSecureConnection()
// 	if err != nil {
// 		t.Error("error opening secure connection with card: ", err)
// 	}
// 	if cs == nil {
// 		t.Error("received nil commandSet when opening secure connection")
// 	}
// }

func TestSelectAndInitialize(t *testing.T) {
	cs, err := Connect()
	if err != nil {
		t.Error("unable to connect to card: ", err)
	}
	//card should start in uninitialized state
	_, _, cardInitialized, err := cs.Select()
	if err != nil {
		t.Error("SELECT failed: ", err)
	}
	if cardInitialized != false {
		t.Error("cardInitialized should be false, but was: ", cardInitialized)
	}

	err = cs.Pair()
	if err != nil {
		t.Error("unable to pair: ", err)
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		t.Error("unable to open secure channel: ", err)
	}

	testPin := "111111"
	err = cs.Init(testPin)
	if err != nil {
		t.Error("could not set pin")
	}
	instanceUID, cardPubKey, cardInitialized, err := cs.Select()
	if err != nil {
		t.Error("could not select initialized card: ", err)
	}
	if cardInitialized != true {
		t.Error("card should be initialized")
	}
	if instanceUID == nil {
		t.Error("instanceUID was nil")
	}
	if cardPubKey == nil {
		t.Error("cardPubKey was nil")
	}
	log.Debugf("InstanceUID: % X", instanceUID)
	log.Debugf("cardPubKey: % X", cardPubKey)
}
