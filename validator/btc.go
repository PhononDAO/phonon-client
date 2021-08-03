package validator

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/GridPlus/phonon-client/model"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"

	log "github.com/sirupsen/logrus"
)

type BTCValidator struct {
	bclient *bcoinClient
}

const transactionRequestLimit int = 6

type bcoinClient struct {
	url       string
	authtoken string
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
	}
}

// Validate returns true if the balance associated with the public key
// on the bitcoin phonon is greater than or equal to the balance stated in 
// the phonon using as many known address generation functions as possible. 
func (b *BTCValidator) Validate(phonon *model.Phonon) (bool, error) {
	// get the public key of the phonon
	key := phonon.PubKey

	// turn it into an address
	addresses, err := pubKeyToAddress(key)
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

func pubKeyToAddress(key *ecdsa.PublicKey) ([]string, error) {
	// TODO: compute addresses for all possible (reasonable) scripts for P2SH addresses
	btcpubkey := btcec.PublicKey{
		Curve: key.Curve,
		X:     key.X,
		Y:     key.Y,
	}
	// something feels wrong about serializing the pubkey just to unserialize it, but hopefully this all gets optimized out so it doesnt matter anyway
	// as far as I can tell, the second argument to this call isn't used for creating an address
	pubKeyUncompressed, err := btcutil.NewAddressPubKey(btcpubkey.SerializeUncompressed(), &chaincfg.MainNetParams)
	if err != nil {
		log.Debug("Error generating address from public key")
		return []string{}, nil
	}

	pubKeyHybrid, err := btcutil.NewAddressPubKey(btcpubkey.SerializeHybrid(), &chaincfg.MainNetParams)
	if err != nil {
		log.Debug("Error generating address from public key")
		return []string{}, nil
	}

	pubKeyCompressed, err := btcutil.NewAddressPubKey(btcpubkey.SerializeCompressed(), &chaincfg.MainNetParams)
	if err != nil {
		log.Debug("Error generating address from public key")
		return []string{}, nil
	}

	return []string{
		pubKeyCompressed.EncodeAddress(),
		pubKeyUncompressed.EncodeAddress(),
		pubKeyHybrid.EncodeAddress(),
	}, nil
}

func (b *BTCValidator) getBalance(addresses []string) (int64, error) {
	fmt.Println("getting balance")
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
		fmt.Println(runningTotal)
		for _, input := range transaction.Inputs {
			for _, address := range addresses {
				if input.Coin.Address == address {
					fmt.Println(fmt.Sprintf("running total: %d, subtracting %d", runningTotal, input.Coin.Value))
					runningTotal -= input.Coin.Value
				}
			}
		}
		for _, output := range transaction.Outputs {
			for _, address := range addresses {
				if output.Address == address {
					fmt.Println(fmt.Sprintf("running total: %d, adding %d", runningTotal, output.Value))
					runningTotal += output.Value
				}
			}
		}
	}
	fmt.Println(runningTotal)
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
			fmt.Println(len(ret))
			// Add limit parameters to url
			url := fmt.Sprintf("%s/tx/address/%s?limit=%d&after=%s", bc.url, address, transactionRequestLimit, ret[len(ret)-1].Hash)
			listPart, err = bc.getTransactionList(ctx, url)
			if err != nil {
				return nil, err
			}
			newret := append(ret, listPart...)
			ret = newret
			fmt.Println(len(ret))
		}
	}
	return ret, nil
}

func (bc *bcoinClient) getTransactionList(ctx context.Context, url string) (transactionList, error) {
	fmt.Println(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Debug("Unable to create request to bcoin api")
		return nil, err
	}

	if bc.authtoken != "" {
		req.SetBasicAuth("x", bc.authtoken)

	}
	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
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
		Value   int64  `json:"value"`
		Script  string `json:"script"`
		Address string `json:"address"`
	} `json:"outputs"`
	Locktime      int    `json:"locktime"`
	Hex           string `json:"hex"`
	Confirmations int    `json:"confirmations"`
}
