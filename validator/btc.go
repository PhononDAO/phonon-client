package validator

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/GridPlus/phonon-client/model"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"

	log "github.com/sirupsen/logrus"
)

var ErrPhononCompromised error = errors.New("Transaction with phonon as sender detected")

type BTCValidator struct {
	bclient *bcoinClient
}

const transactionRequestLimit int = 100

type bcoinClient struct {
	url       string
	authtoken string
	client    http.Client
}

func NewBTCValidator(c *bcoinClient) *BTCValidator {
	return &BTCValidator{
		bclient: c,
	}
}

func NewClient(url string, authToken string) *bcoinClient {
	return &bcoinClient{
		url:       url,
		authtoken: authToken,
		client:    http.Client{},
	}
}

// Validate returns true if the balance associated with the public key
// on the bitcoin phonon is greater than or equal to the balance stated in
// the phonon using as many known address generation functions as reasonable.
// Currently: P2SH script and P2PKH addresses.
func (b *BTCValidator) Validate(phonon *model.Phonon) (bool, error) {
	// get the public key of the phonon
	key := phonon.PubKey

	// turn it into an address
	addresses, err := pubKeyToAddresses(key)
	if err != nil {
		return false, err
	}

	// get balance of address
	balance, err := b.getBalance(addresses)
	if err != nil {
		return false, err
	}

	if balance == 0 {
		return false, nil
	}

	return true, nil
}

func pubKeyToAddresses(key *ecdsa.PublicKey) ([]string, error) {
	btcpubkey := btcec.PublicKey{
		Curve: key.Curve,
		X:     key.X,
		Y:     key.Y,
	}
	var ret = []string{}

	serializationFunctions := []func() []byte{
		btcpubkey.SerializeCompressed,
		btcpubkey.SerializeUncompressed,
		btcpubkey.SerializeHybrid,
	}

	for _, x := range serializationFunctions {
		k, err := btcutil.NewAddressPubKey(x(), &chaincfg.MainNetParams)
		if err != nil {
			log.Debug("Error Generating Address From Public Key")
			return []string{}, err
		}
		ret = append(ret, k.EncodeAddress())

		witnessKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(x()), &chaincfg.MainNetParams)
		if err != nil {
			log.Debug("Error Generating Address Witness From Public Key")
			return []string{}, err

		}
		script, err := txscript.PayToAddrScript(witnessKeyHash)
		if err != nil {
			log.Debug("Error Generating transaction script from witness hash")
			return []string{}, err

		}
		addrScriptHash, err := btcutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
		if err != nil {
			log.Debug("Error Generating Address From PayToAddressScript")
			return []string{}, err

		}

		ret = append(ret, addrScriptHash.EncodeAddress())
	}
	return ret, nil
}

func (b *BTCValidator) getBalance(addresses []string) (int64, error) {
	//get transactions
	transactions, err := b.bclient.GetTransactions(context.Background(), addresses)
	//aggregate transactions into a running balance
	balance, err := aggregateTransactions(transactions, addresses)
	if err != nil {
		return 0, err
	}
	log.Debug("Balance retrieved:", balance)
	return balance, nil
}

func aggregateTransactions(txl transactionList, addresses []string) (int64, error) {
	var runningTotal int64 = 0
	for _, transaction := range txl {
		for _, input := range transaction.Inputs {
			for _, address := range addresses {
				if input.Coin.Address == address {
					return 0, ErrPhononCompromised
				}
			}
		}
		for _, output := range transaction.Outputs {
			for _, address := range addresses {
				if output.Address == address {
					runningTotal += output.Value
				}
			}
		}
	}
	return runningTotal, nil
}

func (bc *bcoinClient) GetTransactions(ctx context.Context, addresses []string) (transactionList, error) {
	var ret transactionList
	for _, address := range addresses {
		url := fmt.Sprintf("%s/tx/address/%s?limit=%d", bc.url, address, transactionRequestLimit)
		var listPart transactionList
		listPart, err := bc.getTransactionList(ctx, url)
		if err != nil {
			return nil, err
		}
		ret = append(ret, listPart...)
		// As long as we are getting a full list, keep checking for more and adding them to the list
		for len(listPart) == transactionRequestLimit {
			// Add limit parameters to url
			url := fmt.Sprintf("%s/tx/address/%s?limit=%d&after=%s", bc.url, address, transactionRequestLimit, ret[len(ret)-1].Hash)
			listPart, err = bc.getTransactionList(ctx, url)
			if err != nil {
				return nil, err
			}
			newret := append(ret, listPart...)
			ret = newret
		}
	}
	return ret, nil
}

func (bc *bcoinClient) getTransactionList(ctx context.Context, url string) (transactionList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Debug("Unable to create request to bcoin api")
		return nil, err
	}

	if bc.authtoken != "" {
		req.SetBasicAuth("x", bc.authtoken)

	}
	resp, err := bc.client.Do(req)
	if err != nil {
		log.Debug("Error making request to bcoin")
		return nil, err
	}
	var ret = transactionList{}
	retBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug("Unable to read response from bcoin api")
		return nil, err
	}

	err = json.Unmarshal(retBytes, &ret)
	if err != nil {
		log.Debug("Unable to unmarshal Json response from bcoin")
		return nil, err
	}
	return ret, nil
}

type transactionList []struct {
	Hash    string  `json:"hash"`
	Inputs  Inputs  `json:"inputs"`
	Outputs Outputs `json:"outputs"`
}

type Inputs []struct {
	Coin Coin `json:"coin"`
}

type Outputs []output

type output struct {
	Value   int64  `json:"value"`
	Address string `json:"address"`
}

type Coin struct {
	Value   int64  `json:"value"`
	Address string `json:"address"`
}
