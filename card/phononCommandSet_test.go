package card

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/GridPlus/phonon-client/model"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{})

	cs, err := Connect()
	if err != nil {
		fmt.Println(err)
		return
	}
	_, _, initialized, err := cs.Select()
	if err != nil {
		fmt.Println(err)
		return
	}
	testPin := "111111"
	if !initialized {
		err = cs.Init(testPin)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
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

func TestSelect(t *testing.T) {
	cs, err := Connect()
	instanceUID, cardPubKey, cardInitialized, err := cs.Select()
	if err != nil {
		t.Error("could not select initialized card: ", err)
		return
	}
	if cardInitialized != true {
		t.Error("card should be initialized")
		return
	}
	if instanceUID == nil {
		t.Error("instanceUID was nil")
		return
	}
	if cardPubKey == nil {
		t.Error("cardPubKey was nil")
		return
	}
	log.Debugf("InstanceUID: % X", instanceUID)
	log.Debugf("cardPubKey: % X", cardPubKey)
}

func TestOpenSecureConnection(t *testing.T) {
	_, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
}

func TestCreateSetAndListPhonons(t *testing.T) {
	cs, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
	type phononDescription struct {
		currencyType model.CurrencyType
		value        float32
	}
	phononTable := []phononDescription{
		{model.Bitcoin, 1},
		{model.Bitcoin, 0.00000001},
		{model.Bitcoin, 99999999},
		{model.Ethereum, 0.000000000000000001},
		{model.Ethereum, 999999999999999999},
		{model.Ethereum, 1},
	}

	//TODO: pass different filters into this function
	type phononFilter struct {
		currencyType model.CurrencyType
		lessThanValue float32
		greaterThanValue float32
	}

	var createdPhonons []model.Phonon
	for _, description := range phononTable {
		keyIndex, pubKey, err := cs.CreatePhonon()
		if err != nil {
			t.Error("err creating test phonon: ", err)
			return
		}
		//track created to review after listing to check that we get out exactly what we put in
		createdPhonons = append(createdPhonons, model.Phonon{
			KeyIndex:     keyIndex,
			PubKey:       pubKey,
			Value:        description.value,
			CurrencyType: description.currencyType})

		err = cs.SetDescriptor(keyIndex, description.currencyType, description.value)
		if err != nil {
			t.Error("err setting test phonon descriptor: ", err)
			return
		}
	}

	//TODO: wrap up as list function, and pass different lists
	receivedPhonons, err := cs.ListPhonons(model.Unspecified, 0, 0)
	if err != nil {
		t.Error("err listing all phonons: ", err)
		return
	}
	// fmt.Print("received phonons: ", receivedPhonons)
	expectedPhononCount := len(createdPhonons)
	var matchedPhononCount int
	for _, received := range receivedPhonons {
		received.PubKey, err = cs.GetPhononPubKey(uint16(received.KeyIndex))
		if err != nil {
			t.Error("could not get phonon pubkey: ", err)
		}
		fmt.Printf("%+v\n", received)
		for _, created := range createdPhonons {
			fmt.Printf("createdPubKey: % X\n", created.PubKey)
			//Todo figure out why this isn't matching
			if received.PubKey.Equal(created.PubKey) {
				matchedPhononCount += 1
				fmt.Printf("received: %+v\n", received)
				fmt.Printf("created: %+v\n", created)
				if !cmp.Equal(received, created) {
					t.Error("phonons with equal pubkeys had different values: ")
					t.Errorf("received: %+v\n", received)
					t.Errorf("created: %+v\n", created)
				}
			}
		}
	}
	if expectedPhononCount > matchedPhononCount {
		t.Errorf("expected %v received phonons to match list but only %v were found", expectedPhononCount, matchedPhononCount)
	}
}

func comparePhononList([]{model.CurrencyType, float32, float32})) {
	for
}
