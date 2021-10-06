//go:generate stringer -type=CurrencyType

package model

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/util"
)

type Phonon struct {
	KeyIndex     uint16
	PubKey       *ecdsa.PublicKey
	Value        float32
	CurrencyType CurrencyType
}

func (p *Phonon) String() string {
	return fmt.Sprintf("KeyIndex: %v, Value: %f, currencyType: %v, PubKey: %v\n", p.KeyIndex, p.Value, p.CurrencyType, util.ECDSAPubKeyToHexString(p.PubKey))
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

type CounterpartyPhononCard interface {
	GetCertificate() (cert.CardCertificate, error)
	CardPair(initPairingData []byte) (cardPairData []byte, err error)
	CardPair2(cardPairData []byte) (cardPairData2 []byte, err error)
	FinalizeCardPair(cardPair2Data []byte) error
	ReceivePhonons(phononTransfer []byte) error
	RequestPhonons(phonons []Phonon) (phononTransfer []byte, err error)
	GenerateInvoice() (invoiceData []byte, err error)
	ReceiveInvoice(invoiceData []byte) error
}
