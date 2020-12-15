package model

type Phonon struct {
	ID         []byte
	Descriptor []Tlv
}

type Tlv struct {
	Tag   string
	Value byte
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
