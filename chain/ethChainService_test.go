package chain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/GridPlus/phonon-client/model"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var simEVM *backends.SimulatedBackend
var genesisKey *ecdsa.PrivateKey
var genesisAcct *bind.TransactOpts

func getSimEVM() (*backends.SimulatedBackend, error) {
	var err error
	if simEVM != nil {
		return simEVM, nil
	}
	genesisKey, _ = crypto.GenerateKey()
	genesisAcct, err = bind.NewKeyedTransactorWithChainID(genesisKey, big.NewInt(1337))
	if err != nil {
		return nil, err
	}
	genesisValue, _ := big.NewInt(0).SetString("1000000000000000000", 0)
	simEVM = backends.NewSimulatedBackend(core.GenesisAlloc{
		genesisAcct.From: {Balance: genesisValue},
	}, 8000000)

	return simEVM, nil
}

func fundSimEVM(currencyType model.CurrencyType) (*model.Phonon, error) {
	sim, err := getSimEVM()
	if err != nil {
		return nil, err
	}

	eth, err := NewEthChainService()
	if err != nil {
		return nil, err
	}

	testChainID := uint32(1337)
	eth.cl = sim
	eth.clChainID = testChainID

	ctx := context.Background()

	nonce, _, _, err := eth.fetchPreTransactionInfo(ctx, genesisAcct.From)
	if err != nil {
		return nil, err
	}

	fixedGasPrice := big.NewInt(875000000)
	phononValue := big.NewInt(10000000000000000)

	senderPrivKey, _ := crypto.GenerateKey()

	pubKey, err := model.NewPhononPubKey(crypto.FromECDSAPub(&senderPrivKey.PublicKey), model.Secp256k1)
	if err != nil {
		return nil, err
	}

	p := &model.Phonon{
		KeyIndex:     1,
		PubKey:       pubKey,
		CurrencyType: currencyType,
		ChainID:      int(testChainID),
	}
	p.Address, err = eth.DeriveAddress(p)
	if err != nil {
		return nil, err
	}
	_, err = eth.submitLegacyTransaction(ctx, nonce,
		big.NewInt(int64(eth.clChainID)),
		common.HexToAddress(p.Address),
		phononValue,
		eth.gasLimit,
		fixedGasPrice,
		genesisKey)
	if err != nil {
		return nil, err
	}

	//Wait for the transaction to be mined
	sim.Commit()

	return p, nil
}

func TestEthChainServiceValidate(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	fund, err := fundSimEVM(model.Ethereum)
	if err != nil {
		t.Error("unable to fund simEVM. err: ", err)
	}

	fund2, err := fundSimEVM(model.Bitcoin)
	if err != nil {
		t.Error("unable to fund simEVM. err: ", err)
	}

	phonons := []*model.Phonon{fund, fund2}

	validator := NewMultiAssetValidator()
	valid, err := validator.Validate(phonons)
	if err != nil {
		t.Error("unable to validate phonons. err: ", err)
	}

	for _, r := range valid {
		if r.P.CurrencyType == model.Bitcoin {
			assert.False(t, r.Valid)
		}

		if r.P.CurrencyType == model.Ethereum {
			assert.True(t, r.Valid)
		}
	}
}
