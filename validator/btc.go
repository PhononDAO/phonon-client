package validator

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/GridPlus/phonon-client/model"
	// "github.com/btcsuite/btcutil"

	log "github.com/sirupsen/logrus"
)

type BTCValidator struct{
	bclient *bcoinClient
}

func (b *BTCValidator) Validate(phonon *model.Phonon) (bool,error){
	// get the public key of the phonon
	key := phonon.PubKey

	// turn it into an address
	address, err := pubKeyToAddress(key)
	if err != nil{
		return false, err
	}
	
	// get balance of address
	balance, err := b.getBalance(address)
	if err != nil{
		return false, err	
	}

	if balance == 0{
		return false, nil	
	}

	return true, nil
}

func pubKeyToAddress(key *ecdsa.PublicKey)(string, error){
	// todo use the btcutil package to do this
	return "",nil
}

func (b *BTCValidator)getBalance(address string)(int64, error){
	//get transactions
	transactions, err := b.bclient.GetTransactions(context.Background(), address)

	//aggregate transactions into a running balance
	balance, err := aggregateTransactions(transactions, address)
	if err != nil{
		return 0, err
	}
	log.Debug("Balance retrieved:", balance)
	return balance, nil
}

func aggregateTransactions(txl transactionList, address string)(int64, error){
	var runningTotal int64 = 0
	for _, transaction := range txl{
		for _, input := range transaction.Inputs{
			if input.Coin.Address == address{
				runningTotal -= input.Coin.Value
			}
		}
		for _, output := range transaction.Outputs{
			if output.Address == address{
				runningTotal += output.Value
			}
		}
	}
	return runningTotal, nil
}

type bcoinClient struct {
	url string
	authtoken string
}

func (bc *bcoinClient)GetTransactions(ctx context.Context, address string)(transactionList, error){
	req, err := http.NewRequestWithContext(ctx,http.MethodGet,bc.url, nil)
	if err != nil{
		log.Debug("Unable to create request to bcoin api")
		return nil, err
	}

	req.SetBasicAuth("x",bc.authtoken)
	
	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil{
		log.Debug("Error making request to bcoin")
		return nil, err
	}

	var ret =  transactionList{}
	retBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Debug("Unable to read response from bcoin api")
		return nil, err
	}

	err = json.Unmarshal(retBytes,&ret)
	if err != nil{
		log.Debug("Unable to unmarshal Json response from bcoin")
		return nil, err
	}

	return ret, nil
}

type transactionList []struct {
	Hash        string `json:"hash"`
	WitnessHash string `json:"witnessHash"`
	Fee         int    `json:"fee"`
	Rate        int    `json:"rate"`
	Mtime       int    `json:"mtime"`
	Height      int    `json:"height"`
	Block       string `json:"block"`
	Time        int    `json:"time"`
	Index       int    `json:"index"`
	Version     int    `json:"version"`
	Inputs      []struct {
		Prevout struct {
			Hash  string `json:"hash"`
			Index int    `json:"index"`
		} `json:"prevout"`
		Script   string `json:"script"`
		Witness  string `json:"witness"`
		Sequence int64  `json:"sequence"`
		Coin     struct {
			Version  int    `json:"version"`
			Height   int    `json:"height"`
			Value    int64  `json:"value"`
			Script   string `json:"script"`
			Address  string `json:"address"`
			Coinbase bool   `json:"coinbase"`
		} `json:"coin"`
	} `json:"inputs"`
	Outputs []struct {
		Value   int64    `json:"value"`
		Script  string `json:"script"`
		Address string `json:"address"`
	} `json:"outputs"`
	Locktime      int    `json:"locktime"`
	Hex           string `json:"hex"`
	Confirmations int    `json:"confirmations"`
}
