package card

import (
	"github.com/GridPlus/phonon-client/model"
)

type localCounterParty struct {
	*Session
}

func NewLocalCounterParty(session *Session) *localCounterParty {
	return &localCounterParty{
		session,
	}
}

func (lcp *localCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	return lcp.cs.CardPair(initPairingData)
}

func (lcp *localCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	return lcp.cs.CardPair2(cardPairData)
}

func (lcp *localCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	return lcp.cs.FinalizeCardPair(cardPair2Data)
}

func (lcp *localCounterParty) ReceivePhonons(phononTransfer []byte) error {
	err := lcp.ReceivePhonons(phononTransfer)
	if err != nil {
		return err
	}
	return nil
}

func (lcp *localCounterParty) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error) {
	//TODO implement
	return nil, nil
}
