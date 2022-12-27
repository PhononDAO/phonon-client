package validator

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/btcec"
)

// Helper type for table testing
type transactionsAndValues struct {
	transactions  transactionList
	addresses     []string
	expectedTotal int64
}

func TestPubKeyToAddresses(t *testing.T) {
	pubkey := "03110c89d71731d603059f919e1670cd335cb915bb7a27b56a667ee057a2e78f3e"

	h, err := hex.DecodeString(pubkey)
	if err != nil {
		t.Error("Unable to decode test public key to hex")
	}

	k, err := btcec.ParsePubKey(h, btcec.S256())
	if err != nil {
		t.Error("Unable to parse public key into btcec.pubkey")
	}

	res, err := pubKeyToAddresses(k.ToECDSA())
	if err != nil {
		t.Error("Received error from PubkeyTo Address")
	}

	expected := []string{
		"1AtZ1U2d2SrW2V8A2Eqicx67zRSDeYYu5k",
		"3EesGzvBgme1o4kB2oFvRnJ9BH3R9c8Uqr",
		"1GAb3tibSSpnXMb5Af3VzPfj3956Xgzewy",
		"349HmcWpNNkGBhEtZq9yFVJrVmtARbzy2d",
		"1FQgJWXuFXiXQ51r1EjkzLyLrYedJ2cXH9",
		"353jJx2n9TcTDrL64cv3H8kRfZhww7Sxwk",
	}

	if !reflect.DeepEqual(res, expected) {
		t.Errorf("Expected results of pubkey to address to be \n%s, but were \n%s", expected, res)
	}

}

func TestCompromisedPhononTransactions(t *testing.T) {
	res, err := aggregateTransactions(transactionListCompromisedPhonon, []string{"target"})
	fmt.Println(res, err)
	if res != 0 || err != ErrPhononCompromised {
		t.Errorf("Expected Compromised Phonon error and zero balance returned, recieved error: %s, and balance of %d", err.Error(), res)
	}
}

func TestAggregateTransactions(t *testing.T) {
	var tv []transactionsAndValues
	tv = append(tv, []transactionsAndValues{
		{
			transactions: list1,
			addresses: []string{
				"phononAddress",
			},
			expectedTotal: int64(50),
		},
		{
			transactions: list2,
			addresses: []string{
				"phononAddress",
			},
			expectedTotal: int64(49),
		},
		{
			transactions: list3,
			addresses: []string{
				"phononAddress",
				"phononAddress2",
				"phononAddress3",
			},
			expectedTotal: int64(149),
		},
	}...,
	)
	for _, x := range tv {
		bal, err := aggregateTransactions(x.transactions, x.addresses)
		if err != nil {
			t.Errorf("Unable to aggregate transactions for transaction: %+v\nerror; %+v", x.transactions, err.Error())
		}
		if bal != x.expectedTotal {
			t.Errorf("Expected balance of %d, got balance %d with input of:\n%+v\nfor addresses:%+v", bal, x.expectedTotal, x.transactions, x.addresses)
		}

	}

}

// Example data for comporomised phonon transaction list
var transactionListCompromisedPhonon transactionList = transactionList{
	{
		Hash:   "NoNeedHere",
		Inputs: Inputs{{Coin: Coin{Value: 100, Address: "target"}}},
		Outputs: []output{
			{
				Value:   int64(50),
				Address: "fakeAddress2",
			},
			{
				Value:   int64(50),
				Address: "fakeAddress3",
			},
		},
	},
}

// Data for phonon testing
var list1 transactionList = transactionList{
	{
		Hash:   "NoNeedHere",
		Inputs: Inputs{{Coin: Coin{Value: 100, Address: "fakeaddress1"}}},
		Outputs: []output{
			{
				Value:   int64(50),
				Address: "fakeAddress2",
			},
			{
				Value:   int64(50),
				Address: "phononAddress",
			},
		},
	},
}
var list2 transactionList = transactionList{
	{
		Hash:   "NoNeedHere",
		Inputs: Inputs{{Coin: Coin{Value: 100, Address: "fakeaddress1"}}},
		Outputs: []output{
			{
				Value:   int64(51),
				Address: "fakeAddress2",
			},
			{
				Value:   int64(49),
				Address: "phononAddress",
			},
		},
	},
}
var list3 transactionList = transactionList{
	{
		Hash:   "NoNeedHere",
		Inputs: Inputs{{Coin: Coin{Value: 100, Address: "fakeaddress1"}}},
		Outputs: []output{
			{
				Value:   int64(51),
				Address: "fakeAddress2",
			},
			{
				Value:   int64(49),
				Address: "phononAddress",
			},
		},
	},
	{
		Hash:   "NoNeedHere",
		Inputs: Inputs{{Coin: Coin{Value: 100, Address: "fakeaddress1"}}},
		Outputs: []output{
			{
				Value:   int64(51),
				Address: "phononAddress2",
			},
			{
				Value:   int64(49),
				Address: "phononAddress",
			},
		},
	},
}
