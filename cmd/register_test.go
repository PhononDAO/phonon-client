package cmd

import (
	"fmt"
	"github.com/GridPlus/phonon-client/util"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func TestRegisterPost(t *testing.T) {
	privKey, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal("could not generate key. err: ", err)
		return
	}

	output, err := registerCard(util.CardIDFromPubKey(&privKey.PublicKey))
	if err != nil {
		t.Fatal("registerCard call failed. err: ", err)
	}
	fmt.Println(string(output))
}
