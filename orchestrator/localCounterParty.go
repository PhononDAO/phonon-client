package orchestrator

import (
	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
)

type localCounterParty struct {
	s *card.Session
}

func NewLocalCounterParty(session *card.Session) *localCounterParty {
	return &localCounterParty{
		s: session,
	}
}

func (lcp *localCounterParty) GetCertificate() (cert.CardCertificate, error) {
	return lcp.s.Cert, nil
}

func (lcp *localCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	return lcp.s.CardPair(initPairingData)
}

func (lcp *localCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	return lcp.s.CardPair2(cardPairData)
}

//TODO: figure out how this state should actually be tracked
func (lcp *localCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	return lcp.s.FinalizeCardPair(cardPair2Data)
}

func (lcp *localCounterParty) ReceivePhonons(phononTransfer []byte) error {
	err := lcp.s.ReceivePhonons(phononTransfer)
	if err != nil {
		return err
	}
	return nil
}

func (lcp *localCounterParty) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error) {
	//TODO implement
	return nil, nil
}

func (lcp *localCounterParty) GenerateInvoice() (invoiceData []byte, err error) {
	return lcp.s.GenerateInvoice()
}

func (lcp *localCounterParty) ReceiveInvoice(invoiceData []byte) error {
	return lcp.s.ReceiveInvoice(invoiceData)
}
