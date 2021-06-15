package card

import (
	"github.com/GridPlus/phonon-client/model"
)

type localCounterParty struct {
	s *Session
}

func NewLocalCounterParty(session *Session) *localCounterParty {
	return &localCounterParty{
		s: session,
	}
}

func (lcp *localCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	return lcp.s.cs.CardPair(initPairingData)
}

func (lcp *localCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	return lcp.s.cs.CardPair2(cardPairData)
}

func (lcp *localCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	return lcp.s.cs.FinalizeCardPair(cardPair2Data)
}

func (lcp *localCounterParty) SendPhonons(phononTransfer []byte) error {
	//TODO implement
	return nil
}

func (lcp *localCounterParty) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error) {
	//TODO implement
	return nil, nil
}
