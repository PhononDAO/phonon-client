package model

import "crypto/ecdsa"

type Phonon struct {
	KeyIndex     int
	PubKey       *ecdsa.PublicKey
	Value        float32
	CurrencyType CurrencyType
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
