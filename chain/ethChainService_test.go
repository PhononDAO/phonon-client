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

func getSimEVM() (simEVM *backends.SimulatedBackend, genesisKey *ecdsa.PrivateKey, genesisAcct *bind.TransactOpts, err error) {
	if simEVM != nil {
		return simEVM, nil, nil, nil
	}
	genesisKey, _ = crypto.GenerateKey()
	genesisAcct, err = bind.NewKeyedTransactorWithChainID(genesisKey, big.NewInt(1337))
	if err != nil {
		return nil, nil, nil, err
	}
	genesisValue, _ := big.NewInt(0).SetString("1000000000000000000", 0)
	simEVM = backends.NewSimulatedBackend(core.GenesisAlloc{
		genesisAcct.From: {Balance: genesisValue},
	}, 8000000)

	return simEVM, genesisKey, genesisAcct, nil
}

func generatePubKey() (model.PhononPubKey, error) {
	senderPrivKey, _ := crypto.GenerateKey()

	pubKey, err := model.NewPhononPubKey(crypto.FromECDSAPub(&senderPrivKey.PublicKey), model.Secp256k1)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func initSimEthChainSrv(chainID int, sim *backends.SimulatedBackend) (*EthChainService, error) {
	ethChainSrv, err := NewEthChainService()
	if err != nil {
		return nil, err
	}

	ethChainSrv.cl = sim

	return ethChainSrv, err
}

func createEthPhonon(desc *model.Phonon) (*model.Phonon, error) {
	pubKey, err := generatePubKey()
	if err != nil {
		return nil, err
	}

	singlePhonon := &model.Phonon{
		KeyIndex:     desc.KeyIndex,
		PubKey:       pubKey,
		CurrencyType: desc.CurrencyType,
		ChainID:      desc.ChainID,
		CurveType:    desc.CurveType,
		Denomination: desc.Denomination,
	}

	return singlePhonon, nil
}

func fundEthPhonon(desc *model.Phonon, ethChainSrv *EthChainService, sim *backends.SimulatedBackend, genesisKey *ecdsa.PrivateKey, genesisAcct *bind.TransactOpts) (*model.Phonon, *EthChainService, error) {
	ctx := context.Background()
	nonce, _, _, err := ethChainSrv.fetchPreTransactionInfo(ctx, genesisAcct.From)
	if err != nil {
		return nil, nil, err
	}

	fixedGasPrice := big.NewInt(875000000)
	phononValue, err := model.NewDenomination(desc.Denomination.Value())
	if err != nil {
		return nil, nil, err
	}

	desc.Address, err = ethChainSrv.DeriveAddress(desc)
	if err != nil {
		return nil, nil, err
	}
	_, err = ethChainSrv.submitLegacyTransaction(ctx, nonce,
		big.NewInt(int64(ethChainSrv.clChainID)),
		common.HexToAddress(desc.Address),
		phononValue.Value(),
		ethChainSrv.gasLimit,
		fixedGasPrice,
		genesisKey)
	if err != nil {
		return nil, nil, err
	}

	//Wait for the transaction to be mined
	sim.Commit()

	return desc, ethChainSrv, nil
}

func TestValidate(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	sim, key, acct, err := getSimEVM()
	if err != nil {
		t.Error("unable to get simEVM. err: ", err)
	}

	testChainID := 1337

	ethChainSrv, err := initSimEthChainSrv(testChainID, sim)
	if err != nil {
		t.Error("unable to init simEVM. err: ", err)
	}

	value, err := model.NewDenomination(big.NewInt(10000000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	highValue, err := model.NewDenomination(big.NewInt(9900000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	lowValue, err := model.NewDenomination(big.NewInt(1000000000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	phonons := []*model.Phonon{}
	for i := 0; i < 10; i++ {
		phonon, err := createEthPhonon(&model.Phonon{
			KeyIndex:     model.PhononKeyIndex(i),
			CurrencyType: model.Ethereum,
			ChainID:      testChainID,
			CurveType:    model.Secp256k1,
			Denomination: value,
		})
		if err != nil {
			t.Error("unable to create phonon. err: ", err)
		}

		if i == 2 {
			phonon.ChainID = 9
		}

		ethChainSrv.clChainID = uint32(phonon.ChainID)

		phonon, ethChainSrv, err = fundEthPhonon(phonon, ethChainSrv, sim, key, acct)
		if err != nil {
			t.Error("unable to fund phonon. err: ", err)
		}

		if i == 0 {
			phonon.Denomination = lowValue
		}

		if i == 3 {
			phonon.CurrencyType = model.Bitcoin
		}

		if i == 4 {
			phonon.Denomination = highValue
		}

		phonons = append(phonons, phonon)
	}

	valid, err := ethChainSrv.Validate(phonons)
	if err != nil {
		t.Error("unable to validate phonons. err: ", err)
	}

	for _, r := range valid {
		if r.P.KeyIndex >= 5 {
			t.Log("validating a true phonon")
			assert.Equal(t, r.Valid, true)
			assert.Equal(t, r.Err, nil)
		}

		if r.P.KeyIndex == 0 {
			t.Logf("validating an invalid phonon with an invalid denomination: %s", r.P.Denomination)
			assert.Equal(t, r.Valid, false)
			assert.EqualError(t, r.Err, model.ErrBalanceTooLow.Error())
		}

		// if r.P.KeyIndex == 2 {
		// 	t.Logf("validating an invalid phonon with an unsupported chain id: %d", r.P.ChainID)
		// 	assert.Equal(t, r.Valid, false)
		// 	assert.EqualError(t, r.Err, model.ErrBalanceTooLow.Error())
		// }

		if r.P.KeyIndex == 3 {
			t.Logf("validating an invalid phonon with an unsupported currency type: %s", r.P.CurrencyType)
			assert.Equal(t, r.Valid, false)
			assert.EqualError(t, r.Err, model.ErrUnsupportedCurrency.Error())
		}

		if r.P.KeyIndex == 4 {
			t.Logf("validating an invalid phonon with an invalid denomination: %s", r.P.Denomination)
			assert.Equal(t, r.Valid, false)
			assert.EqualError(t, r.Err, model.ErrBalanceTooHigh.Error())
		}
	}
}
