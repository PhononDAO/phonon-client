package chain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/GridPlus/phonon-client/model"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

var ErrUntrustworthy = errors.New("phonon balance is less than what is stated in denomination value")
var ErrUnsupportedCurrency = errors.New("unsupported currency type")

type EthChainInterface interface {
	bind.ContractTransactor
	ethereum.ChainStateReader
}
type EthChainService struct {
	gasLimit  uint64
	cl        EthChainInterface //*ethclient.Client // //bind.ContractTransactor
	clChainID uint32
}

func NewEthChainService() (*EthChainService, error) {
	ethchainSrv := &EthChainService{
		gasLimit: uint64(21000), //Setting to default magic value for now
	}
	log.Debugf("successfully loaded EthChainServiceConfig: %+v", ethchainSrv)

	return ethchainSrv, nil
}

func (eth *EthChainService) Validate(proposal []*model.Phonon) (result []*model.AssetValidationResult, err error) {
	if len(proposal) == 0 {
		return nil, model.ErrEmptyProposal
	}

	for _, p := range proposal {
		err := eth.dialRPCNode(p.ChainID)
		if err != nil {
			result = append(result, &model.AssetValidationResult{
				P:   p,
				Err: err,
			})
		}

		balance, err := eth.cl.BalanceAt(context.Background(), common.HexToAddress(p.Address), nil)
		if err != nil {
			log.Error("could not fetch on chain Phonon value: ", err)
			result = append(result, &model.AssetValidationResult{
				P:   p,
				Err: err,
			})
		}

		if balance.Cmp(p.Denomination.Value()) < 0 {
			log.Error("phonon balance is less than the denomination value")
			result = append(result, &model.AssetValidationResult{
				P:   p,
				Err: ErrUntrustworthy,
			})
		}

		result = append(result, &model.AssetValidationResult{
			P:     p,
			Valid: true,
			Err:   nil,
		})
	}

	return result, nil
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

	//Send the transaction through the ETH client
	err = eth.cl.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Error("error sending transaction: ", err)
		return signedTx, err
	}
	log.Debug("sent redeem transaction")
	return signedTx, nil
}

func (eth *EthChainService) DeriveAddress(p *model.Phonon) (address string, err error) {
	eccPubKey, err := model.PhononPubKeyToECDSA(p.PubKey)
	if err != nil {
		return "", err
	}
	return ethcrypto.PubkeyToAddress(*eccPubKey).Hex(), nil
}

// dialRPCNode establishes a connection to the proper RPC node based on the chainID
func (eth *EthChainService) dialRPCNode(chainID int) (err error) {
	log.Debugf("ethChainID: %v, chainID: %v\n", eth.clChainID, chainID)
	var RPCEndpoint string
	//If chainID is already set, correct RPC node is already connected
	if eth.clChainID != 0 && eth.clChainID == uint32(chainID) {
		log.Debug("eth chainID already set to ", chainID)
		return nil
	}
	switch chainID {
	// case 1: //Mainnet
	// 	//untested
	// 	RPCEndpoint = "https://eth-mainnet.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 5: // Goerli
		RPCEndpoint = "https://eth-goerli.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 42: //Kovan
		RPCEndpoint = "https://poa-kovan.gateway.pokt.network/v1/lb/621e9e234e140e003a32b8ba"
	case 97: // Binance Testnet
		RPCEndpoint = "https://data-seed-prebsc-1-s1.binance.org:8545"
	case 43114: // Avalanche Fuji (C-Chain)
		RPCEndpoint = "https://api.avax-test.network/ext/bc/C/rpc"
	case 80001: // Mumbai (Polygon)
		RPCEndpoint = "https://rpc-mumbai.maticvigil.com"
	case 4002: // Fantom testnet
		RPCEndpoint = "https://rpc.testnet.fantom.network"
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
	eth.clChainID = uint32(chainID)
	log.Trace("eth chain ID set to ", chainID)
	return nil
}
