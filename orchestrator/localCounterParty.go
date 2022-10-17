package orchestrator

import (
	"errors"
	"fmt"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
)

type localCounterParty struct {
	// remote session
	counterSession *Session
	localSession   *Session
	pairingStatus  model.RemotePairingStatus
}

func init() {
	connectedCardsAndLCPSessions = make(map[*Session]*localCounterParty)
}

var connectedCardsAndLCPSessions map[*Session]*localCounterParty

func (lcp *localCounterParty) ConnectToCard(cardID string) error {
	counterparty := globalTerminal.SessionFromID(cardID)
	if counterparty == nil {
		return errors.New("counterparty card not found")
	}
	// connect the cards one way
	lcp.counterSession = counterparty
	// connect the other direction
	connectedCardsAndLCPSessions[counterparty].counterSession = lcp.localSession
	return nil
}

func (lcp *localCounterParty) GetCertificate() (*cert.CardCertificate, error) {
	return lcp.counterSession.GetCertificate()
}

func (lcp *localCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	return lcp.counterSession.CardPair(initPairingData)
}

func (lcp *localCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	lcp.pairingStatus = model.StatusPaired
	return lcp.counterSession.CardPair2(cardPairData)
}

func (lcp *localCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	lcp.pairingStatus = model.StatusPaired
	return lcp.counterSession.FinalizeCardPair(cardPair2Data)
}

func (lcp *localCounterParty) ReceivePhonons(phononTransfer []byte) error {
	return lcp.counterSession.ReceivePhonons(phononTransfer)
}

func (lcp *localCounterParty) GenerateInvoice() (invoiceData []byte, err error) {
	return lcp.counterSession.GenerateInvoice()
}

func (lcp *localCounterParty) ReceiveInvoice(invoiceData []byte) error {
	return lcp.counterSession.ReceiveInvoice(invoiceData)
}

func (lcp *localCounterParty) VerifyPaired() error {
	if lcp.pairingStatus == model.StatusPaired {
		return nil
	} else {
		return fmt.Errorf("not paired to local counterparty")
	}
}

func (lcp *localCounterParty) PairingStatus() model.RemotePairingStatus {
	return lcp.pairingStatus
}
