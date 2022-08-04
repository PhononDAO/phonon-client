package chain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

//Composite interface supporting all needed EVM RPC calls
type EthChainInterface interface {
	bind.ContractTransactor
	ethereum.ChainStateReader
}
type EthChainService struct {
	gasLimit  uint64
	cl        EthChainInterface //*ethclient.Client // //bind.ContractTransactor
	clChainID int
}

func NewEthChainService() (*EthChainService, error) {
	ethchainSrv := &EthChainService{
		gasLimit: uint64(21000), //Setting to default magic value for now
	}
	log.Debugf("successfully loaded EthChainServiceConfig: %+v", ethchainSrv)

	return ethchainSrv, nil
}

//Derives an ETH address from a phonon's ECDSA Public Key
func (eth *EthChainService) DeriveAddress(p *model.Phonon) (address string, err error) {
	eccPubKey, err := model.PhononPubKeyToECDSA(p.PubKey)
	if err != nil {
		return "", err
	}
	return ethcrypto.PubkeyToAddress(*eccPubKey).Hex(), nil
}

func (eth *EthChainService) RedeemPhonon(p *model.Phonon, privKey *ecdsa.PrivateKey, redeemAddress string) (transactionData string, err error) {
	err = eth.ValidateRedeemData(p, privKey, redeemAddress)
	if err != nil {
		log.Error("phonon did not contain complete data for redemption: ", err)
		return "", err
	}

	//Will validate that we have a valid RPC endpoint for the given p.ChainID
	err = eth.dialRPCNode(p.ChainID)
	if err != nil {
		return "", err
	}
	ctx := context.Background()

	//Collect on chain details for redeem
	nonce, onChainBalance, suggestedGasPrice, err := eth.fetchPreTransactionInfo(ctx, common.HexToAddress(p.Address))
	if err != nil {
		return "", err
	}
	redeemValue := eth.calcRedemptionValue(onChainBalance, suggestedGasPrice)
	log.Debug("transaction redemption value is: ", redeemValue)

	//If gas would cost more than the value in the phonon, return error
	if suggestedGasPrice.Cmp(onChainBalance) != -1 {
		log.Error("phonon not large enough to pay gas for redemption")
		return "", errors.New("phonon not large enough to pay gas for redemption")
	}

	tx, err := eth.submitLegacyTransaction(ctx, nonce, big.NewInt(int64(p.ChainID)), common.HexToAddress(redeemAddress), redeemValue, eth.gasLimit, suggestedGasPrice, privKey)
	if err != nil {
		return "", err
	}

	//Parse Response
	return tx.Hash().String(), nil
}

//ReconcileRedeemData validates the input data to ensure it contains all that's needed for a successful redemption.
//It will update the phonon data structure with a derived address if necessary
func (eth *EthChainService) ValidateRedeemData(p *model.Phonon, privKey *ecdsa.PrivateKey, redeemAddress string) (err error) {
	eccPubKey, err := model.PhononPubKeyToECDSA(p.PubKey)
	if err != nil {
		return err
	}
	//Check that pubkey listed in metadata matches pubKey derived from phonon's private key
	if !eccPubKey.Equal(privKey.Public()) {
		log.Error("phonon pubkey metadata and pubkey derived from redemption privKey did not match. err: ", err)
		log.Error("metadata pubkey: ", util.ECCPubKeyToHexString(eccPubKey))
		log.Error("privKey derived key: ", util.ECCPubKeyToHexString(&privKey.PublicKey))
		return errors.New("pubkey metadata and redemption private key did not match")
	}

	//Check that fromAddress exists, if not derive it
	if p.Address == "" {
		p.Address, err = eth.DeriveAddress(p)
		if err != nil {
			log.Error("unable to derive source address for redemption: ", err)
			return err
		}
	}

	//Check that redeemAddress is valid
	//Just checks for correct address length, works with or without 0x prefix
	valid := common.IsHexAddress(redeemAddress)
	if !valid {
		return errors.New("redeem address invalid")
	}

	return nil
}

//dialRPCNode establishes a connection to the proper RPC node based on the chainID
func (eth *EthChainService) dialRPCNode(chainID int) (err error) {
	log.Debugf("ethChainID: %v, chainID: %v\n", eth.clChainID, chainID)
	var RPCEndpoint string
	//If chainID is already set, correct RPC node is already connected
	if eth.clChainID != 0 && eth.clChainID == chainID {
		return nil
	}
	switch chainID {
	case 1: //Mainnet
		//untested
		RPCEndpoint = "https://eth-mainnet.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 3: //Ropsten
		//untested
		RPCEndpoint = "https://eth-ropsten.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 4: //Rinkeby
		RPCEndpoint = "https://eth-rinkeby.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 42: //Kovan
		RPCEndpoint = "https://poa-kovan.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 97: // Binance
		RPCEndpoint = "https://data-seed-prebsc-1-s1.binance.org:8545"
	case 43113: // Avalanche
		RPCEndpoint = "https://api.avax-test.network/ext/bc/C/rpc"
	case 80001: // Polygon
		RPCEndpoint = "https://rpc-mumbai.maticvigil.com"
	case 1337: //Local Ganache
		RPCEndpoint = "HTTP://127.0.0.1:8545"
	default:
		log.Debug("unsupported eth chainID requested")
		return errors.New("eth chainID unsupported")
	}
	eth.cl, err = ethclient.Dial(RPCEndpoint)
	if err != nil {
		log.Errorf("could not dial eth chain provider at endpoint %v: %v\n", RPCEndpoint, err)
		return err
	}

	//If connection succeeded, set currently configured chainID
	eth.clChainID = chainID
	log.Debug("eth chain ID set to ", chainID)
	return nil
}

func (eth *EthChainService) fetchPreTransactionInfo(ctx context.Context, fromAddress common.Address) (nonce uint64, balance *big.Int, suggestedGas *big.Int, err error) {
	nonce, err = eth.cl.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Error("could not fetch pending nonce for eth account: ", err)
		return 0, nil, nil, err
	}
	log.Debug("pending nonce: ", nonce)
	//Check actual balance of phonon
	balance, err = eth.cl.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Error("could not fetch on chain Phonon value: ", err)
		return 0, nil, nil, err
	}
	log.Debug("on chain balance: ", balance)
	suggestedGasPrice, err := eth.cl.SuggestGasPrice(ctx)
	if err != nil {
		log.Error("error fetching suggested gas price: ", err)
		return 0, nil, nil, err
	}
	log.Debug("suggest gas price is: ", suggestedGasPrice)
	return nonce, balance, suggestedGasPrice, nil
}

func (eth *EthChainService) calcRedemptionValue(balance *big.Int, gasPrice *big.Int) *big.Int {
	valueMinusGas := big.NewInt(0)
	estimatedGasCost := big.NewInt(0)
	gasLimit := int(eth.gasLimit)
	return valueMinusGas.Sub(balance, estimatedGasCost.Mul(gasPrice, big.NewInt(int64(gasLimit))))
}

func (eth *EthChainService) checkBalance(ctx context.Context, address string) (balance *big.Int, err error) {
	balance, err = eth.cl.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		log.Error("could not fetch on chain Phonon value: ", err)
		return nil, err
	}
	return balance, nil
}

func (eth *EthChainService) submitLegacyTransaction(ctx context.Context, nonce uint64, chainID *big.Int, redeemAddress common.Address, redeemValue *big.Int, gasLimit uint64, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	//Submit transaction
	//build transaction payload
	tx := types.NewTransaction(nonce, redeemAddress, redeemValue, gasLimit, gasPrice, nil)
	//Sign it
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
	if err != nil {
		log.Error("error forming signed transaction: ", err)
		return signedTx, err
	}

	balance, err := eth.checkBalance(ctx, redeemAddress.Hex())
	if err != nil {
		log.Error("error checking balance: ", err)
		return signedTx, err
	}

	if redeemValue.Cmp(eth.calcRedemptionValue(balance, gasPrice)) < 0 {
		log.Error("balance is insufficient to cover redemption value")
		return signedTx, errors.New("balance is insufficient to cover redemption value")
	}

	//Send the transaction through the ETH client
	err = eth.cl.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Error("error sending transaction: ", err)
		return signedTx, err
	}
	log.Debug("sent redeem transaction")
	return signedTx, nil
}
