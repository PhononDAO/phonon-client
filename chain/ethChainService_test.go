package chain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/GridPlus/phonon-client/model"
	// "github.com/GridPlus/phonon-client/util"
	// "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/core/types"

	// "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
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

//TestEthChainServiceRedeem smoke tests the basic redeem funtionality using a ganache backend.
//Ganache must be stood up manually and this test must be hand edited with valid keys to function.
func TestEthChainServiceRedeem(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	sim, err := getSimEVM()
	if err != nil {
		t.Error("unable to start EVM simulator")
		return
	}
	//Hand Edit Here!
	//Manually change privKeyHex and redeemAddress to values from the ganache backend used for this test
	// privKeyHex := "287f9caac470d6d8c0a921f60f912f81572ebd4aee6f91c41fdd20f950b27d1f"
	// redeemAddress := "0x18579269D059CD91581A01C2C3d70B16940c1BA7"

	//Generate sender account
	senderPrivKey, _ := crypto.GenerateKey()

	//Generate destination account
	redeemPrivKey, _ := crypto.GenerateKey()
	redeemAddress := crypto.PubkeyToAddress(redeemPrivKey.PublicKey)

	eth, err := NewEthChainService()
	if err != nil {
		t.Error(err)
	}

	//Manually substitute simulated backend for usual RPC client
	testChainID := 1337
	eth.cl = sim
	eth.clChainID = testChainID

	pubKey, err := model.NewPhononPubKey(crypto.FromECDSAPub(&senderPrivKey.PublicKey), model.Secp256k1)
	if err != nil {
		t.Fatal("could not construct pubKey: ", err)
	}

	p := &model.Phonon{
		KeyIndex:     1,
		PubKey:       pubKey,
		CurrencyType: model.Ethereum,
		ChainID:      testChainID,
	}
	p.Address, err = eth.DeriveAddress(p)
	if err != nil {
		t.Error(err)
	}
	ctx := context.Background()
	//Fund phonon with sim ETH from genesis account
	nonce, _, _, err := eth.fetchPreTransactionInfo(ctx, genesisAcct.From)
	if err != nil {
		t.Error("could not fetch banker transaction info", err)
		return
	}
	fixedGasPrice := big.NewInt(875000000)
	phononValue := big.NewInt(10000000000000000)
	_, err = eth.submitLegacyTransaction(ctx, nonce,
		big.NewInt(int64(testChainID)),
		common.HexToAddress(p.Address),
		phononValue,
		eth.gasLimit,
		fixedGasPrice,
		genesisKey)
	if err != nil {
		t.Error("unable to submit simulated phonon deposit transaction. err: ", err)
	}

	//Commit funding transaction
	sim.Commit()

	//Testing redeem function
	//Redeem to redeemAddress
	_, err = eth.RedeemPhonon(p, senderPrivKey, redeemAddress.Hex())
	if err != nil {
		t.Error("error redeeming phonon. err: ", err)
	}
	//Mine the pending transaction
	sim.Commit()

	//Check that the redeem succeeded
	resultBalance, err := eth.cl.BalanceAt(context.Background(), redeemAddress, nil)
	if err != nil {
		t.Error("could not check balance of phonon redeem address. err: ", err)
	}

	//TODO: boundary cases with expected errors

	//Calculate how much the gas fee * gas limit cost to redeem
	expectedBalance := big.NewInt(0)
	calculatedRedeemGasPrice, _ := big.NewInt(0).SetString("766199219", 10) //Taken from log of redeem results
	redeemGasPaid := big.NewInt(0).Mul(calculatedRedeemGasPrice, big.NewInt(21000))

	//TODO: Test a range of phonon values
	expectedBalance = expectedBalance.Sub(phononValue, redeemGasPaid)
	if expectedBalance.Cmp(resultBalance) != 0 {
		t.Error("expected balance was incorrect")
		t.Errorf("expectedBalance: %v, resultBalance: %v, phononValue: %v, redeemGasPaid: %v\n", expectedBalance, resultBalance, phononValue, redeemGasPaid)
	}
	//Check for correct balance output
	t.Log("resultBalance was: ", resultBalance)
}
