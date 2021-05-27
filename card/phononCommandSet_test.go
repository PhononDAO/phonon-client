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

var testPin string = "111111"

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	log.SetLevel(log.DebugLevel)
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
	if !initialized {
		err = cs.Init(testPin)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	os.Exit(m.Run())
}

//SELECT
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

//PAIR
//OPEN_SECURE_CHANNEL
//MUTUAL_AUTH
func TestOpenSecureConnection(t *testing.T) {
	_, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
}

//VERIFY_PIN
//CREATE_PHONON
//SET_DESCRIPTOR
//GET_PHONON_PUB_KEY
//LIST_PHONONS
func TestCreateSetAndListPhonons(t *testing.T) {
	cs, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
	err = cs.VerifyPIN(testPin)
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
		currencyType        model.CurrencyType
		lessThanValue       float32
		greaterThanValue    float32
		expectedPhononCount int
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
	fmt.Printf("createdPhonons: %+v", createdPhonons)

	filters := []phononFilter{
		{
			currencyType:        model.Unspecified,
			lessThanValue:       0,
			greaterThanValue:    0,
			expectedPhononCount: 6,
		},
		{
			currencyType:        model.Bitcoin,
			lessThanValue:       0,
			greaterThanValue:    0,
			expectedPhononCount: 3,
		},
		{
			currencyType:        model.Ethereum,
			lessThanValue:       0,
			greaterThanValue:    0,
			expectedPhononCount: 3,
		},
	}

	for _, f := range filters {
		//TODO: wrap up as list function, and pass different lists
		receivedPhonons, err := cs.ListPhonons(f.currencyType, f.lessThanValue, f.greaterThanValue)
		if err != nil {
			t.Error("err listing all phonons: ", err)
			return
		}
		fmt.Println("len of the received phonons: ", len(receivedPhonons))
		// fmt.Print("received phonons: ", receivedPhonons)
		var matchedPhononCount int

		for _, received := range receivedPhonons {
			received.PubKey, err = cs.GetPhononPubKey(uint16(received.KeyIndex))
			if err != nil {
				t.Error("could not get phonon pubkey: ", err)
				return
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
		if f.expectedPhononCount != matchedPhononCount {
			t.Errorf("expected %v received phonons to match list but only %v were found", f.expectedPhononCount, matchedPhononCount)
		}
	}
}

// DESTROY_PHONON
func TestDestroyPhonon(t *testing.T) {
	cs, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
	if err = cs.VerifyPIN(testPin); err != nil {
		t.Error(err)
		return
	}
	keyIndex, createdPubKey, err := cs.CreatePhonon()
	if err != nil {
		t.Error(err)
		return
	}
	err = cs.SetDescriptor(keyIndex, model.Ethereum, .578)
	if err != nil {
		t.Error(err)
		return
	}
	privKey, err := cs.DestroyPhonon(keyIndex)
	if err != nil {
		t.Error(err)
		return
	}

	resultPubKey := privKey.PublicKey
	if !createdPubKey.Equal(&resultPubKey) {
		t.Errorf("createdPubKey: %+v", createdPubKey)
		t.Errorf("derivedPubKey: %+v", resultPubKey)
		t.Error("derived pubKey from destroyed phonon was not equivalent to created PubKey")
		t.Error("privKey from destroy: % X", append(privKey.X.Bytes(), privKey.Y.Bytes()...))
		t.Errorf("derived result: % X\n", append(resultPubKey.X.Bytes(), resultPubKey.Y.Bytes()...))
		t.Errorf("created pubKey: % X\n", append(createdPubKey.X.Bytes(), createdPubKey.Y.Bytes()...))
	}
}

// TODO
//Pairing + Send/Receive cycle
// SEND_PHONONS
// SET_RECV_LIST
// RECV_PHONONS
// TRANSACTION_ACK
// CARD_PAIR
// CARD_PAIR_2
// FINALIZE_CARD_PAIRING
// IDENTIFY_CARD

// func TestTransactionAck(t *testing.T) {
// 	cs, err := OpenSecureConnection()
// 	if err != nil {
// 		t.Error("could not open secure connection: ", err)
// 		return
// 	}
// 	err = cs.VerifyPIN(testPin)
// 	if err != nil {
// 		t.Error("could not verify pin: ", err)
// 		return
// 	}
// 	keyIndex, _, err := cs.CreatePhonon()
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	err = cs.SetDescriptor(keyIndex, model.Bitcoin, 1)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	_, err = cs.SendPhonons([]uint16{keyIndex}, false)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	err = cs.TransactionAck([]uint16{keyIndex})
// 	if err != nil {
// 		t.Error("error in transaction ack: ", err)
// 		return
// 	}
// }
