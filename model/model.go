//go:generate stringer -type=CurrencyType

package model

import "crypto/ecdsa"

type Phonon struct {
	KeyIndex     uint16
	PubKey       *ecdsa.PublicKey
	Value        float32
	CurrencyType CurrencyType
}

type CurrencyType uint16

const (
	Unspecified CurrencyType = 0x0000
	Bitcoin     CurrencyType = 0x0001
	Ethereum    CurrencyType = 0x0002
)

type CryptoChain byte

const (
	testnet CryptoChain = iota
)

//Key: denomination
//value: quantity of the associated denomination
type CoinList map[int]int
