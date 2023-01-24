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

func fundEthPhonon(phonon *model.Phonon, ethChainSrv *EthChainService, sim *backends.SimulatedBackend, genesisKey *ecdsa.PrivateKey, genesisAcct *bind.TransactOpts) (*model.Phonon, *EthChainService, error) {
	ctx := context.Background()
	nonce, _, _, err := ethChainSrv.fetchPreTransactionInfo(ctx, genesisAcct.From)
	if err != nil {
		return nil, nil, err
	}

	fixedGasPrice := big.NewInt(875000000)

	phononValue, err := model.NewDenomination(phonon.Denomination.Value())
	if err != nil {
		return nil, nil, err
	}

	phonon.Address, err = ethChainSrv.DeriveAddress(phonon)
	if err != nil {
		return nil, nil, err
	}
	_, err = ethChainSrv.submitLegacyTransaction(ctx, nonce,
		big.NewInt(int64(ethChainSrv.clChainID)),
		common.HexToAddress(phonon.Address),
		phononValue.Value(),
		ethChainSrv.gasLimit,
		fixedGasPrice,
		genesisKey)
	if err != nil {
		return nil, nil, err
	}

	//Wait for the transaction to be mined
	sim.Commit()

	return phonon, ethChainSrv, nil
}

func TestValidate(t *testing.T) {
	type validateTest struct {
		p     *model.Phonon
		valid bool
		err   error
	}

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

	// 1e zeros
	value, err := model.NewDenomination(big.NewInt(100000000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	// 14 zeros
	lowValue, err := model.NewDenomination(big.NewInt(900000000000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	// 8 zeros
	highValue, err := model.NewDenomination(big.NewInt(100000000))
	if err != nil {
		t.Error("unable to add denomination value. err: ", err)
	}

	vt := []validateTest{
		{
			p: &model.Phonon{
				KeyIndex:     0,
				Denomination: value,
				ChainID:      testChainID,
				CurveType:    model.Secp256k1,
				CurrencyType: model.Ethereum,
			},
			valid: true,
		},
		{
			p: &model.Phonon{
				KeyIndex:     1,
				Denomination: value,
				ChainID:      testChainID,
				CurveType:    model.Secp256k1,
				CurrencyType: model.Bitcoin,
			},
			valid: false,
			err:   model.ErrUnsupportedCurrency,
		},
		{
			p: &model.Phonon{
				KeyIndex:     2,
				Denomination: value,
				ChainID:      testChainID,
				CurveType:    model.Secp256k1,
				CurrencyType: model.Ethereum,
			},
			valid: false,
			err:   model.ErrBalanceTooLow,
		},
		{
			p: &model.Phonon{
				KeyIndex:     3,
				Denomination: value,
				ChainID:      testChainID,
				CurveType:    model.Secp256k1,
				CurrencyType: model.Ethereum,
			},
			valid: false,
			err:   model.ErrBalanceTooHigh,
		},
	}

	phononV := []*model.Phonon{}
	for _, phonon := range vt {
		phonon.p.PubKey, err = generatePubKey()
		if err != nil {
			t.Error("unable to generate pub key. err: ", err)
		}

		phononsToValidate, ethChainSrv, err := fundEthPhonon(phonon.p, ethChainSrv, sim, key, acct)
		if err != nil {
			t.Error("unable to fund phonon. err: ", err)
		}

		phononV = append(phononV, phononsToValidate)

		for i := range phononV {
			if phononV[i].KeyIndex == 3 {
				phononV[i].Denomination = highValue
			}

			if phononV[i].KeyIndex == 2 {
				phononV[i].Denomination = lowValue
			}

			validationResult, err := ethChainSrv.Validate(phononV)
			if err != nil {
				t.Error("unable to validate phonons. err: ", err)
			}

			if validationResult[i].Valid != vt[i].valid {
				t.Errorf("expected validation result to be %v, got %v for index %v", vt[i].valid, validationResult[i].Valid, vt[i].p.KeyIndex)
			}

			if validationResult[i].Err != vt[i].err {
				t.Errorf("expected validation error to be %v, got %v for index %v", vt[i].err, validationResult[i].Err, vt[i].p.KeyIndex)
			}
		}
	}
}
