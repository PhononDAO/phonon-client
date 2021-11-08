package model

import (
	"crypto/ecdsa"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/util"
)

type PhononCard interface {
	Select() (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error)
	Pair() (*cert.CardCertificate, error)
	OpenSecureChannel() error
	Init(pin string) error
	IdentifyCard(nonce []byte) (cardPubKey *ecdsa.PublicKey, cardSig *util.ECDSASignature, err error)
	VerifyPIN(pin string) error
	ChangePIN(pin string) error
	CreatePhonon() (keyIndex uint16, pubKey *ecdsa.PublicKey, err error)
	SetDescriptor(phonon *Phonon) error
	ListPhonons(currencyType CurrencyType, lessThanValue uint64, greaterThanValue uint64) ([]*Phonon, error)
	GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error)
	DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error)
	SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets []byte, err error)
	ReceivePhonons(phononTransfer []byte) error
	SetReceiveList(phononPubKeys []*ecdsa.PublicKey) error
	TransactionAck(keyIndices []uint16) error
	InitCardPairing(receiverCertificate cert.CardCertificate) (initPairingData []byte, err error)
	CardPair(initPairingData []byte) (cardPairData []byte, err error)
	CardPair2(cardPairData []byte) (cardPair2Data []byte, err error)
	FinalizeCardPair(cardPair2Data []byte) (err error)
	InstallCertificate(signKeyFunc func([]byte) ([]byte, error)) (err error)
	GenerateInvoice() (invoiceData []byte, err error)
	ReceiveInvoice(invoiceData []byte) (err error)
}
