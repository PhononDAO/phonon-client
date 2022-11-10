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
	VerifyPaired() error
	PairingStatus() RemotePairingStatus
	ConnectToCard(string) error

	RecieveProposedTransaction(phononProposalPacket []byte) (err error)
	ReceiveTransfer(transferPakcet []byte) error
	CancelTransfer()
}

type RemotePairingStatus int

const (
	StatusUnconnected RemotePairingStatus = iota
	StatusConnectedToBridge
	StatusConnectedToCard
	StatusCardPair1Complete
	StatusCardPair2Complete
	StatusPaired
)
