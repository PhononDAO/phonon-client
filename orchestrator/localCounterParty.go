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
	err := lcp.counterSession.ReceivePhonons(phononTransfer)
	if err != nil {
		return err
	}
	return nil
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

func (lcp *localCounterParty) RecieveProposedTransaction(phononProposalPacket []byte) (err error) {
	if lcp.pairingStatus != model.StatusPaired {
		return ErrCardNotPairedToCard
	}
	_, err = lcp.counterSession.ReceiveTransferProposal(phononProposalPacket)
	if err != nil {
		return err
	}
	// todo: receive phonons from this call and put them in queue to be verified
	return nil
}
func (lcp *localCounterParty) ReceiveTransfer(transferPacket []byte) error {
	if lcp.pairingStatus != model.StatusPaired {
		return ErrCardNotPairedToCard
	}
	return lcp.counterSession.ReceivePhonons(transferPacket)
}
func (lcp *localCounterParty) CancelTransfer() {
	_ = lcp.counterSession.cs.CancelTransfer()
}
