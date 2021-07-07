package validator

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"testing"

	"github.com/GridPlus/phonon-client/model"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

func TestEthValidation(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	var seed string
	for i := 0; i < 40; i++ {
		seed += "B"
	}
	staticSeed := strings.NewReader(seed)
	key, err := ecdsa.GenerateKey(ethcrypto.S256(), staticSeed)
	if err != nil {
		t.Error("could not generate static test key")
	}
	fmt.Printf("static PubKey: % X\n", append(key.PublicKey.X.Bytes(), key.PublicKey.Y.Bytes()...))
	ethAddress := ethcrypto.PubkeyToAddress(key.PublicKey)
	fmt.Println("static ETH adress: ", ethAddress)

	samplePhonon := &model.Phonon{
		PubKey:       &key.PublicKey,
		CurrencyType: model.Ethereum,
		ChainID:      int(Rinkeby),
		Value:        3,
	}

	ev := EthValidator{}
	valid, err := ev.Validate(samplePhonon)
	if err != nil {
		t.Error("error validating test phonon")
		t.Error(err)
	}
	if !valid {
		t.Error("phonon invalid")
	}
}
