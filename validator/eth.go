package validator

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/GridPlus/phonon-client/model"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	log "github.com/sirupsen/logrus"
)

type EthChainID int

//chain IDs based on EIP-155: https://eips.ethereum.org/EIPS/eip-155#list-of-chain-ids
const (
	Mainnet EthChainID = 1 //Not implemented
	Ropsten EthChainID = 3 //Not implemented
	Rinkeby EthChainID = 4
	Goerli  EthChainID = 5  //Not implemented
	Kovan   EthChainID = 42 //Not implemented
)

const ()

//Infura based ethereum phonon validator

type EthValidator struct {
	c *ethclient.Client
}

//Infura based validator for now, but underlying client could be swapped out via config
func NewEthValidator(infuraURL string) (*EthValidator, error) {
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		return nil, err
	}
	return &EthValidator{c: client}, nil
}

func connectInfuraEndpoint(chainID EthChainID) (*ethclient.Client, error) {
	var infuraURL string
	switch chainID {
	case Rinkeby:
		infuraURL = "https://rinkeby.infura.io"
	}
	//Collect Infura Key
	infuraAPIKey := os.Getenv("INFURA_API_KEY")
	if infuraAPIKey == "" {
		return nil, errors.New("infura api key not found")
	}
	client, err := ethclient.Dial(infuraURL + "/v3/" + infuraAPIKey)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (e *EthValidator) Validate(phonon *model.Phonon) (valid bool, err error) {
	//Derive ETH Address
	if phonon.PubKey == nil {
		return false, ErrMissingPubKey
	}
	ethAddress := ethcrypto.PubkeyToAddress(*phonon.PubKey)
	log.Debug("derived eth address: ", ethAddress)

	//Connect ETH client to network by chainID
	client, err := connectInfuraEndpoint(EthChainID(phonon.ChainID))
	if err != nil {
		return false, err
	}

	//check account is externally owned
	//TODO: decide what context to use
	contractCode, err := client.CodeAt(context.Background(), ethAddress, nil)
	if err != nil {
		log.Debug("error requesting client code: ", err)
		return false, err
	}
	log.Debugf("contract code: %v", contractCode)
	if len(contractCode) > 0 {
		log.Debug("ETH account is not externally owned")
		return false, nil
	}

	//Make HTTP Infura Request for balance
	balance, err := client.BalanceAt(context.Background(), ethAddress, nil)
	if err != nil {
		log.Debug("error requesting ETH balance ", err)
		return false, err
	}

	log.Debug("balance retrieved: ", balance)
	//Interpret results
	// if balance < (*big.Int)(phonon.Value) {
	// 	return false, nil
	// }
	//Return answer
	return true, nil
}
