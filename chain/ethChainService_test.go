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
)

func getSimEVM() (simEVM *backends.SimulatedBackend, genesisKey *ecdsa.PrivateKey, genesisAcct *bind.TransactOpts, err error) {
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
	ethChainSrv.clChainID = uint32(chainID)

	return ethChainSrv, err
}

func fundEthPhonon(phonons []*model.Phonon, ethChainSrv *EthChainService, sim *backends.SimulatedBackend, genesisKey *ecdsa.PrivateKey, genesisAcct *bind.TransactOpts) ([]*model.Phonon, *EthChainService, error) {
	ctx := context.Background()
	nonce, _, _, err := ethChainSrv.fetchPreTransactionInfo(ctx, genesisAcct.From)
	if err != nil {
		return nil, nil, err
	}

	fixedGasPrice := big.NewInt(875000000)

	for _, p := range phonons {
		phononValue, err := model.NewDenomination(p.Denomination.Value())
		if err != nil {
			return nil, nil, err
		}

		p.Address, err = ethChainSrv.DeriveAddress(p)
		if err != nil {
			return nil, nil, err
		}
		_, err = ethChainSrv.submitLegacyTransaction(ctx, nonce,
			big.NewInt(int64(ethChainSrv.clChainID)),
			common.HexToAddress(p.Address),
			phononValue.Value(),
			ethChainSrv.gasLimit,
			fixedGasPrice,
			genesisKey)
		if err != nil {
			return nil, nil, err
		}
	}

	//Wait for the transaction to be mined
	sim.Commit()

	return phonons, ethChainSrv, nil
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

	// 13 zeros
	value, err := model.NewDenomination(big.NewInt(10000000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	// 12 zeros
	highValue, err := model.NewDenomination(big.NewInt(9900000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	// 15 zeros
	lowValue, err := model.NewDenomination(big.NewInt(1000000000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	pubKey, err := generatePubKey()
	if err != nil {
		t.Error("unable to generate pub key. err: ", err)
	}

	phonons := []*model.Phonon{
		{
			KeyIndex:     0,
			Denomination: value,
			CurrencyType: model.Ethereum,
			PubKey:       pubKey,
			CurveType:    model.Secp256k1,
			ChainID:      testChainID,
		},
		{
			KeyIndex:     1,
			Denomination: lowValue,
			CurrencyType: model.Ethereum,
			PubKey:       pubKey,
			CurveType:    model.Secp256k1,
			ChainID:      testChainID,
		},
		{
			KeyIndex:     2,
			Denomination: highValue,
			CurrencyType: model.Ethereum,
			PubKey:       pubKey,
			CurveType:    model.Secp256k1,
			ChainID:      testChainID,
		},
		{
			KeyIndex:     3,
			Denomination: value,
			CurrencyType: model.Bitcoin,
			PubKey:       pubKey,
			CurveType:    model.Secp256k1,
			ChainID:      testChainID,
		},
	}

	phononsToValidate, ethChainSrv, err := fundEthPhonon(phonons, ethChainSrv, sim, key, acct)
	if err != nil {
		t.Error("unable to fund phonon. err: ", err)
	}

	validationResult, err := ethChainSrv.Validate(phononsToValidate)
	if err != nil {
		t.Error("unable to validate phonons. err: ", err)
	}

	for _, r := range validationResult {
		if r.P.KeyIndex == 0 {
			if !(r.Valid == true && r.Err == nil) {
				t.Error("valid phonon should return true with no error")
			}
		}

		if r.P.KeyIndex == 1 {
			if !(r.Valid == false && r.Err == model.ErrBalanceTooLow) {
				t.Error("invalid phonon with an invalid denomination should return an error")
			}
		}

		if r.P.KeyIndex == 3 {
			if !(r.Valid == false && r.Err == model.ErrUnsupportedCurrency) {
				t.Error("invalid phonon with an unsupported currency type should return an error")
			}
		}

		if r.P.KeyIndex == 2 {
			if !(r.Valid == false && r.Err == model.ErrBalanceTooHigh) {
				t.Error("invalid phonon with an invalid denomination should return an error")
			}
		}
	}
}
