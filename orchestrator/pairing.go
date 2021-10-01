package orchestrator

import (
	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
)

type Pairing struct {
	s      *card.Session
	remote *remoteSession
	name   string
}

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

	err = p.remote.counterParty.ReceivePhonons(phononTransferPacket)
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
	invoiceData, err := p.remote.counterParty.GenerateInvoice()
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
	remoteCert, err := remoteCard.GetCertificate()
	if err != nil {
		return err
	}
	initPairingData, err := p.s.InitCardPairing(remoteCert)
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
	p.remote = &remoteSession{remoteCard}

	return nil
}
