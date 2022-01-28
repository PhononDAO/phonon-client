package orchestrator

import (
	"fmt"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/session"

	"github.com/GridPlus/phonon-client/model"
)

type localCounterParty struct {
	s             *session.Session
	pairingStatus model.RemotePairingStatus
}

func NewLocalCounterParty(session *session.Session) *localCounterParty {
	return &localCounterParty{
		s:             session,
		pairingStatus: model.StatusConnectedToCard,
	}
}

func (lcp *localCounterParty) GetCertificate() (*cert.CardCertificate, error) {
	return lcp.s.GetCertificate()
}

func (lcp *localCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	return lcp.s.CardPair(initPairingData)
}

func (lcp *localCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	return lcp.s.CardPair2(cardPairData)
}

func (lcp *localCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	lcp.pairingStatus = model.StatusPaired
	return lcp.s.FinalizeCardPair(cardPair2Data)
}

func (lcp *localCounterParty) ReceivePhonons(phononTransfer []byte) error {
	err := lcp.s.ReceivePhonons(phononTransfer)
	if err != nil {
		return err
	}
	return nil
}

func (lcp *localCounterParty) GenerateInvoice() (invoiceData []byte, err error) {
	return lcp.s.GenerateInvoice()
}

func (lcp *localCounterParty) ReceiveInvoice(invoiceData []byte) error {
	return lcp.s.ReceiveInvoice(invoiceData)
}

func (lcp *localCounterParty) VerifyPaired() error {
	if lcp.pairingStatus == model.StatusPaired {
		return nil
	} else {
		return fmt.Errorf("Not paired to local counterparty")
	}
}

func (lcp *localCounterParty) PairingStatus() model.RemotePairingStatus {
	return lcp.pairingStatus
}
