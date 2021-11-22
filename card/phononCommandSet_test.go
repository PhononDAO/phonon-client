package card

import (
	"fmt"

	"testing"

	"github.com/GridPlus/phonon-client/model"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

//Card must be initialized with this pin before integration test suite can run
var testPin string = "111111"

//SELECT
func TestSelect(t *testing.T) {
	cs, err := Connect()
	if err != nil {
		t.Error(err)
		return
	}
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
		value        model.Denomination
	}
	phononTable := []phononDescription{
		{model.Bitcoin, model.Denomination{1, 0}},
		{model.Bitcoin, model.Denomination{1, 8}},
		{model.Bitcoin, model.Denomination{99, 6}},
		{model.Ethereum, model.Denomination{1, 18}},
		{model.Ethereum, model.Denomination{9, 18}},
		{model.Ethereum, model.Denomination{1, 0}},
	}

	type phononFilter struct {
		currencyType        model.CurrencyType
		lessThanValue       uint64
		greaterThanValue    uint64
		expectedPhononCount int
	}

	var createdPhonons []*model.Phonon
	for _, description := range phononTable {
		keyIndex, pubKey, err := cs.CreatePhonon(model.Secp256k1)
		if err != nil {
			t.Error("err creating test phonon: ", err)
			return
		}
		p := &model.Phonon{
			KeyIndex:     keyIndex,
			CurrencyType: description.currencyType,
			Denomination: description.value,
		}
		//track created to review after listing to check that we get out exactly what we put in
		createdPhonons = append(createdPhonons, &model.Phonon{
			KeyIndex:     keyIndex,
			PubKey:       pubKey,
			Denomination: description.value,
			CurrencyType: description.currencyType})

		err = cs.SetDescriptor(p)
		if err != nil {
			t.Error("err setting test phonon descriptor: ", err)
			return
		}
	}
	// fmt.Printf("createdPhonons: %+v", createdPhonons)

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
		// fmt.Printf("listing phonons with filter: %+v\n", f)
		receivedPhonons, err := cs.ListPhonons(f.currencyType, f.lessThanValue, f.greaterThanValue)
		if err != nil {
			t.Error("err listing all phonons: ", err)
			return
		}
		// fmt.Println("len of the received phonons: ", len(receivedPhonons))
		// fmt.Print("received phonons: ", receivedPhonons)
		var matchedPhononCount int

		for _, received := range receivedPhonons {
			received.PubKey, err = cs.GetPhononPubKey(uint16(received.KeyIndex))
			if err != nil {
				t.Errorf("could not get phonon pubkey at index %v: %v\n", received.KeyIndex, err)
				return
			}
			// fmt.Printf("%+v\n", received)
			for _, created := range createdPhonons {
				// fmt.Printf("createdPubKey: % X\n", created.PubKey)
				if received.PubKey.Equal(created.PubKey) {
					matchedPhononCount += 1
					if !cmp.Equal(received, created) {
						t.Error("error: phonons with equal pubkeys had different values: ")
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
	keyIndex, createdPubKey, err := cs.CreatePhonon(model.Secp256k1)
	if err != nil {
		t.Error(err)
		return
	}
	p := &model.Phonon{
		KeyIndex:     keyIndex,
		CurrencyType: model.Ethereum,
		Denomination: model.Denomination{57, 0},
	}
	err = cs.SetDescriptor(p)
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

//DestroyPhonon and reuse keyIndex

//Create maximum number of phonons (256) and list them
func TestFillPhononTable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestFillPhononTable in short mode")
	}
	cs, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
	if err = cs.VerifyPIN(testPin); err != nil {
		t.Error(err)
		return
	}
	initialList, err := cs.ListPhonons(model.Unspecified, 0, 0)
	if err != nil {
		t.Error(err)
		return
	}
	initialCount := len(initialList)
	maxPhononCount := 256
	var createdIndices []uint16
	for i := 0; i < maxPhononCount-initialCount; i++ {
		keyIndex, _, err := cs.CreatePhonon(model.Secp256k1)
		if err != nil {
			t.Error(err)
			return
		}
		createdIndices = append(createdIndices, keyIndex)
	}
	list, err := cs.ListPhonons(model.Unspecified, 0, 0)
	if err != nil {
		t.Error(err)
		return
	}
	//Check that all phonons created were listed
	if len(list) != maxPhononCount {
		t.Error(err)
		return
	}
	//Clean up all phonons before next test
	for _, keyIndex := range createdIndices {
		_, err := cs.DestroyPhonon(keyIndex)
		if err != nil {
			t.Error("unable to delete phonon at keyIndex ", keyIndex)
			t.Error(err)
			return
		}
	}
}

func TestReuseDestroyedIndex(t *testing.T) {
	cs, err := OpenSecureConnection()
	if err != nil {
		t.Error(err)
		return
	}
	if err = cs.VerifyPIN(testPin); err != nil {
		t.Error(err)
		return
	}
	//Create three phonons so we can check reusing an index from the middle, beginning, and end of the list
	keyIndex1, _, err := cs.CreatePhonon(model.Secp256k1)
	if err != nil {
		t.Error(err)
		return
	}
	keyIndex2, _, err := cs.CreatePhonon(model.Secp256k1)
	if err != nil {
		t.Error(err)
		return
	}
	keyIndex3, _, err := cs.CreatePhonon(model.Secp256k1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("created phonons for reuse and destroy check at indices, %v, %v, and %v\n", keyIndex1, keyIndex2, keyIndex3)
	//Check the indices in order middle, last, first to ensure all index positions are properly reused
	DestroyReuseAndCheck := func(keyIndex uint16) error {
		//Destroy and reused the middle index
		_, err = cs.DestroyPhonon(keyIndex)
		if err != nil {
			t.Error(err)
			return err
		}
		//Should be equivalent to index
		reusedIndex, _, err := cs.CreatePhonon(model.Secp256k1)
		if err != nil {
			t.Error(err)
			return err
		}
		if reusedIndex != keyIndex {
			t.Errorf("keyIndex %v not reused correctly\n", keyIndex)
			return err
		}
		return nil
	}
	DestroyReuseAndCheck(keyIndex2)
	DestroyReuseAndCheck(keyIndex3)
	DestroyReuseAndCheck(keyIndex1)

}

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
