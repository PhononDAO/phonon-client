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

func initSimEthChainSrv(chainId uint32, sim *backends.SimulatedBackend) (*EthChainService, error) {
	ethChainSrv, err := NewEthChainService()
	if err != nil {
		return nil, err
	}

	ethChainSrv.cl = sim
	ethChainSrv.clChainID = chainId

	return ethChainSrv, err
}

// read transactions / docs, describes how eth works

func fundEthPhonon(value *big.Int, chainId int, ethChainSrv *EthChainService, sim *backends.SimulatedBackend, genesisKey *ecdsa.PrivateKey, genesisAcct *bind.TransactOpts) (*model.Phonon, *EthChainService, error) {
	ctx := context.Background()
	nonce, _, _, err := ethChainSrv.fetchPreTransactionInfo(ctx, genesisAcct.From)
	if err != nil {
		return nil, nil, err
	}

	fixedGasPrice := big.NewInt(875000000)
	phononValue, err := model.NewDenomination(value)
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := generatePubKey()
	if err != nil {
		return nil, nil, err
	}

	p := &model.Phonon{
		KeyIndex:     1,
		PubKey:       pubKey,
		CurrencyType: model.Ethereum,
		ChainID:      chainId,
		CurveType:    model.Secp256k1,
		Denomination: phononValue,
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

	//Wait for the transaction to be mined
	sim.Commit()

	return p, ethChainSrv, nil
}

func TestValidate(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	sim, key, acct, err := getSimEVM()
	if err != nil {
		t.Error("unable to get simEVM. err: ", err)
	}

	chainID := 1337

	ethChainSrv, err := initSimEthChainSrv(uint32(chainID), sim)
	if err != nil {
		t.Error("unable to init simEVM. err: ", err)
	}

	// should take in an entire phonon def.
	// change facility to check for validity of phonons on chain
	phonon, ethChainSrv, err := fundEthPhonon(big.NewInt(1000000000000000000), chainID, ethChainSrv, sim, key, acct)
	if err != nil {
		t.Error("unable to fund simEVM. err: ", err)
	}

	phonons := []*model.Phonon{phonon}
	valid, err := ethChainSrv.Validate(phonons)
	if err != nil {
		t.Error("unable to validate phonons. err: ", err)
	}

	for _, r := range valid {
		t.Log(r.P.KeyIndex, r.P.CurrencyType, r.Valid, r.Err)
	}
}
