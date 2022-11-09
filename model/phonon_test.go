package model

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestDenominationErrorCheck(t *testing.T) {
	type denomErrTest struct {
		input  *big.Int
		output error
	}

	tt := []denomErrTest{
		{big.NewInt(355), ErrInvalidDenomination},
		{big.NewInt(256), ErrInvalidDenomination},
		{big.NewInt(10010), ErrInvalidDenomination},
	}

	for _, test := range tt {
		_, err := NewDenomination(test.input)
		if err != test.output {
			t.Error("did not receive ErrInvalidDenomination for input: ", test.input)
		}
	}
}

func TestDenominationSetAndPrint(t *testing.T) {
	type denomTest struct {
		input  *big.Int
		output string
	}

	//36 zeros
	reallyBigString := "9000000000000000000000000000000000000"
	var reallyBigInt *big.Int
	reallyBigInt, _ = big.NewInt(0).SetString(reallyBigString, 10)

	tt := []denomTest{
		{big.NewInt(10), "10"},
		{big.NewInt(15), "15"},
		{big.NewInt(199), "199"},
		{big.NewInt(1000), "1000"},
		{big.NewInt(1500), "1500"},
		{big.NewInt(1200000000), "1200000000"},
		{big.NewInt(1000000000000000000), "1000000000000000000"},
		{reallyBigInt, reallyBigString},
	}
	log.SetLevel(log.DebugLevel)
	for _, test := range tt {
		d, err := NewDenomination(test.input)
		if err != nil {
			t.Error(err)
		}
		log.Tracef("d: %+v\n", d)
		if d.String() != test.output {
			t.Error("error value output should be '100000' but was ", d.String())
		}
	}
}

func TestDenominationJSONUnmarshal(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	denomJSON := []byte(`{"Denomination":"1000"}`)

	type denominationStruct struct {
		Denomination Denomination
	}
	result := &denominationStruct{}
	correct := &Denomination{Base: 100, Exponent: 1}
	err := json.Unmarshal(denomJSON, result)
	if err != nil {
		t.Error("error unmarshaling denomination from JSON. err: ", err)
	}
	if result.Denomination.Base != 100 || result.Denomination.Exponent != 1 {
		t.Errorf("denomination did not unmarshal correctly. was %+v, should be %+v\n", result, correct)
	}
}

func TestDenominationJSONMarshal(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	type denominationStruct struct {
		Denomination Denomination
	}
	d, _ := NewDenomination(big.NewInt(1000))
	log.Trace("printed d: ", d)
	input := denominationStruct{
		Denomination: d,
	}
	result, err := json.Marshal(&input)
	if err != nil {
		t.Error("error marshalling denomination: ", err)
	}
	log.Trace("printed result: ", string(result))
	correct := []byte(`{"Denomination":"1000"}`)
	if string(result) != string(correct) {
		t.Errorf("resulting json incorrect. was %v, should be %v\n", string(result), string(correct))
	}
}

// TestMarshalJSON tests Marshalling the interesting fields from the phonon struct
func TestMarshalPhononJSON(t *testing.T) {
	dInt, _ := big.NewInt(0).SetString("1000000000000000", 10) //15 zeros
	d, err := NewDenomination(dInt)
	if err != nil {
		t.Error("error setting denomination: ", d)
	}
	pubKeyBytes, err := hex.DecodeString("041ecfecb19648bb85de8ee4d39b0d06ce5586da71e2e177e94ef98de24edf8eaef57fa76617033d145d7e5dd8b0965148a0825241e7983e0a40421f942492018b")
	if err != nil {
		t.Fatal("error decoding hex pubKey")
	}
	pubKey, err := NewPhononPubKey(pubKeyBytes, Secp256k1)
	if err != nil {
		t.Fatal("error parsing pubKey. err: ", err)
	}
	p := &Phonon{
		KeyIndex:     1,
		PubKey:       pubKey,
		Denomination: d,
		CurrencyType: Ethereum,
		ChainID:      1337,
	}
	JSON, err := json.Marshal(p)
	if err != nil {
		t.Error("could not JSONMarshal phonon: ", err)
	}
	correctJSON := string([]byte(`{"KeyIndex":1,"PubKey":"041ecfecb19648bb85de8ee4d39b0d06ce5586da71e2e177e94ef98de24edf8eaef57fa76617033d145d7e5dd8b0965148a0825241e7983e0a40421f942492018b","Address":"","AddressType":0,"SchemaVersion":0,"ExtendedSchemaVersion":0,"Denomination":"1000000000000000","CurrencyType":2,"ChainID":1337,"CurveType":0}`))

	JSONstring := string(JSON)

	if JSONstring != correctJSON {
		t.Error("marshalled phonon JSON not equal to correct JSON.")
		t.Errorf("was %v\nshould be %v\n", JSONstring, correctJSON)
	}
}

func TestMarshalAndUnmarshalPhonon(t *testing.T) {
	testJSON := []byte(`{"KeyIndex":1,"PubKey":"041ecfecb19648bb85de8ee4d39b0d06ce5586da71e2e177e94ef98de24edf8eaef57fa76617033d145d7e5dd8b0965148a0825241e7983e0a40421f942492018b","Address":"","AddressType":0,"SchemaVersion":0,"ExtendedSchemaVersion":0,"Denomination":"1000000000000000","CurrencyType":2,"ChainID":1337,"CurveType":0}`)

	p := &Phonon{}
	err := json.Unmarshal(testJSON, p)
	if err != nil {
		t.Error("couldn't unmarshal phonon JSON. err: ", err)
	}
	resultJSON, err := json.Marshal(p)
	if err != nil {
		t.Error("couldn't marshal phonon JSON. err: ", err)
	}

	testJSONstring := string(testJSON)
	resultString := string(resultJSON)
	if testJSONstring != resultString {
		t.Error("phonon encode/decode did not match")
		t.Errorf("was %v\n should be %v\n", resultString, testJSONstring)
	}
}
