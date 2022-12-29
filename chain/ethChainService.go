package chain

import (
	"context"
	"errors"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/validator"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

type EthChainInterface interface {
	bind.ContractTransactor
	ethereum.ChainStateReader
}
type EthChainService struct {
	cl        EthChainInterface //*ethclient.Client // //bind.ContractTransactor
	clChainID uint32
}

func NewEthChainService() *EthChainService {
	return &EthChainService{}
}

func (eth *EthChainService) Validate(proposal []*model.Phonon) ([]*validator.AssetValidationResult, error) {
	result := []*validator.AssetValidationResult{}

	for _, p := range proposal {
		err := eth.dialRPCNode(p.ChainID)
		if err != nil {
			result = append(result, &validator.AssetValidationResult{
				P:     p,
				Valid: false,
				Err:   err,
			})
		}

		balance, err := eth.cl.BalanceAt(context.Background(), common.HexToAddress(p.Address), nil)
		if err != nil {
			log.Error("could not fetch on chain Phonon value: ", err)
			result = append(result, &validator.AssetValidationResult{
				P:     p,
				Valid: false,
				Err:   err,
			})
		}

		if balance.Cmp(p.Denomination.Value()) < 0 {
			log.Error("phonon balance is less than the denomination value")
			result = append(result, &validator.AssetValidationResult{
				P:     p,
				Valid: false,
				Err:   errors.New("phonon balance is less than what is stated in denomination value"),
			})
		}

		result = append(result, &validator.AssetValidationResult{
			P:     p,
			Valid: true,
			Err:   nil,
		})
	}

	return result, nil
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
