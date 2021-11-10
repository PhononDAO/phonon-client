package model

import (
	"github.com/GridPlus/phonon-client/cert"
)

type CounterpartyPhononCard interface {
	GetCertificate() (*cert.CardCertificate, error)
	CardPair(initPairingData []byte) (cardPairData []byte, err error)
	CardPair2(cardPairData []byte) (cardPairData2 []byte, err error)
	FinalizeCardPair(cardPair2Data []byte) error
	ReceivePhonons(phononTransfer []byte) error
	RequestPhonons(phonons []Phonon) (phononTransfer []byte, err error)
	GenerateInvoice() (invoiceData []byte, err error)
	ReceiveInvoice(invoiceData []byte) error
}
