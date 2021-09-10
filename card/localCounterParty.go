package card

import (
	"github.com/GridPlus/phonon-client/model"
)

type localCounterParty struct {
	s *Session
}

func NewLocalCounterParty(session *Session) *localCounterParty {
	return &localCounterParty{
		session,
	}
}

func (lcp *localCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	return lcp.s.cs.CardPair(initPairingData)
}

func (lcp *localCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	return lcp.s.cs.CardPair2(cardPairData)
}

//TODO: figure out how this state should actually be tracked
func (lcp *localCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	return lcp.s.cs.FinalizeCardPair(cardPair2Data)
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
	return lcp.s.cs.GenerateInvoice()
}

func (lcp *localCounterParty) ReceiveInvoice(invoiceData []byte) error {
	return lcp.s.cs.ReceiveInvoice(invoiceData)
}
