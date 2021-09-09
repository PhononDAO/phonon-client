package terminal

import (
	"errors"
	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
)

type PhononTerminal struct {
	sessions []*Pairing
}

type Pairing struct {
	s      *card.Session
	remote model.CounterpartyPhononCard
}

var ErrRemoteNotPaired error = errors.New("no remote card paired")

func NewPairing(s *card.Session) *Pairing {
	return &Pairing{
		s: s,
	}
}

func (p *Pairing) SendPhonons(keyIndices []uint16) error {
	if p.remote == nil {
		return ErrRemoteNotPaired
	}
	phononTransferPacket, err := p.s.SendPhonons(keyIndices)
	if err != nil {
		return err
	}

	err = p.remote.ReceivePhonons(phononTransferPacket)
	if err != nil {
		return err
	}
	return nil
}

//Retrieve invoice data from a remote paired card
func (p *Pairing) RetrieveInvoice() error {
	if p.remote == nil {
		return ErrRemoteNotPaired
	}
	invoiceData, err := p.remote.GenerateInvoice()
	if err != nil {
		return err
	}
	err = p.s.ReceiveInvoice(invoiceData)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pairing) PairWithRemoteCard(remoteCard model.CounterpartyPhononCard) error {
	initPairingData, err := p.s.InitCardPairing()
	if err != nil {
		return err
	}
	cardPairData, err := remoteCard.CardPair(initPairingData)
	if err != nil {
		return err
	}
	cardPair2Data, err := p.s.CardPair2(cardPairData)
	if err != nil {
		return err
	}
	err = remoteCard.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	p.s.SetPairing(true)
	p.remote = remoteCard

	return nil
}
