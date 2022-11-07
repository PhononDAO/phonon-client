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
	OpenSecureConnection() error
	Init(pin string) error
	IdentifyCard(nonce []byte) (cardPubKey *ecdsa.PublicKey, cardSig *util.ECDSASignature, err error)
	VerifyPIN(pin string) error
	ChangePIN(pin string) error
	CreatePhonon(curveType CurveType) (keyIndex PhononKeyIndex, pubKey PhononPubKey, err error)
	SetDescriptor(phonon *Phonon) error
	ListPhonons(currencyType CurrencyType, lessThanValue uint64, greaterThanValue uint64, continuation bool) ([]*Phonon, error)
	GetPhononPubKey(keyIndex PhononKeyIndex, crv CurveType) (pubkey PhononPubKey, err error)
	DestroyPhonon(keyIndex PhononKeyIndex) (privKey *ecdsa.PrivateKey, err error)
	SendPhonons(keyIndices []PhononKeyIndex, extendedRequest bool) (transferPhononPackets []byte, err error)
	ReceivePhonons(phononTransfer []byte) error
	SetReceiveList(phononPubKeys []*ecdsa.PublicKey) error
	TransactionAck(keyIndices []PhononKeyIndex) error
	InitCardPairing(receiverCertificate cert.CardCertificate) (initPairingData []byte, err error)
	CardPair(initPairingData []byte) (cardPairData []byte, err error)
	CardPair2(cardPairData []byte) (cardPair2Data []byte, err error)
	FinalizeCardPair(cardPair2Data []byte) (err error)
	LoadCertAuthority(CAPubKey []byte) (err error)
	InstallCertificate(signKeyFunc func([]byte) ([]byte, error)) (err error)
	GenerateInvoice() (invoiceData []byte, err error)
	ReceiveInvoice(invoiceData []byte) (err error)
	SetFriendlyName(name string) error
	GetFriendlyName() (string, error)
	GetAvailableMemory() (persistentMem int, onResetMem int, onDeselectMem int, err error)
	MineNativePhonon(difficulty uint8) (keyIndex PhononKeyIndex, hash []byte, err error)
}
