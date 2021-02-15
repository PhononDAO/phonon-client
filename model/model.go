package model

import (
	"crypto/ecdsa"

	"github.com/GridPlus/phonon-client/chain"
)

type Phonon struct {
	keyIndex     int
	pubKey       *ecdsa.PublicKey
	value        float32
	currencyType chain.CurrencyType
}

type CryptoAsset byte

const (
	Test CryptoAsset = iota
	ETH
	BTC
)

type CryptoChain byte

const (
	testnet CryptoChain = iota
)

//Key: denomination
//value: quantity of the associated denomination
type CoinList map[int]int
