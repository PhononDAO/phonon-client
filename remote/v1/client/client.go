package client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	v1 "github.com/GridPlus/phonon-client/remote/v1"
	"github.com/GridPlus/phonon-client/util"
	"github.com/posener/h2conn"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

type RemoteConnection struct {
	conn                     *h2conn.Conn
	out                      *gob.Encoder
	in                       *gob.Decoder
	remoteCertificate        *cert.CardCertificate
	localCertificate         *cert.CardCertificate
	sessionRequestChan       chan model.SessionRequest
	identifiedWithServerChan chan bool
	identifiedWithServer     bool
	counterpartyNonce        [32]byte
	verified                 bool
	connectedToCardChan      chan bool
	verifyPairedChan         chan string

	//card pairing message channels
	remoteCertificateChan    chan cert.CardCertificate
	remoteIdentityChan       chan []byte
	cardPair1DataChan        chan []byte
	finalizeCardPairDataChan chan []byte
	pairingStatus            model.RemotePairingStatus

	phononAckChan chan bool
}

var ErrTimeout = errors.New("Timeout")

// Requests into the card session
func (c *RemoteConnection) getLocalCertificate() (*cert.CardCertificate, error) {
	req := &model.RequestCertificate{
		Ret: make(chan model.ResponseCertificate),
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	if ret.Err != nil {
		return &cert.CardCertificate{}, ret.Err
	}
	return ret.Payload, nil
}

func (c *RemoteConnection) requestIdentifyCard(payload []byte) (*ecdsa.PublicKey, *util.ECDSASignature, error) {
	req := &model.RequestIdentifyCard{
		Ret:   make(chan model.ResponseIdentifyCard),
		Nonce: payload,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.PubKey, ret.Sig, ret.Err
}

func (c *RemoteConnection) requestCardPair1(payload []byte) ([]byte, error) {
	req := &model.RequestCardPair1{
		Ret:     make(chan model.ResponseCardPair1),
		Payload: payload,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Payload, ret.Err
}

func (c *RemoteConnection) requestFinalizeCardPair(payload []byte) error {
	req := &model.RequestFinalizeCardPair{
		Ret:     make(chan model.ResponseFinalizeCardPair),
		Payload: payload,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func (c *RemoteConnection) requestSetRemoteCard(card model.CounterpartyPhononCard) error {
	req := &model.RequestSetRemote{
		Ret:  make(chan model.ResponseSetRemote),
		Card: card,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func (c *RemoteConnection) requestReceivePhonons(payload []byte) error {
	req := &model.RequestReceivePhonons{
		Ret:     make(chan model.ResponseReceivePhonons),
		Payload: payload,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func (c *RemoteConnection) requestGetName() (string, error) {
	req := &model.RequestGetName{
		Ret: make(chan model.ResponseGetName),
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Name, ret.Err
}

func (c *RemoteConnection) requestPairWithRemote(card model.CounterpartyPhononCard) error {
	req := &model.RequestPairWithRemote{
		Ret:  make(chan model.ResponsePairWithRemote),
		Card: card,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func (c *RemoteConnection) requestSetPaired(status bool) error {
	req := &model.RequestSetPaired{
		Ret:    make(chan model.ResponseSetPaired),
		Status: status,
	}
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func Connect(sessReqChan chan model.SessionRequest, url string, ignoreTLS bool) (*RemoteConnection, error) {
	d := &h2conn.Client{
		Client: &http.Client{
			Transport: &http2.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreTLS}},
		},
	}

	conn, resp, err := d.Connect(context.Background(), url) //url)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to remote server %e,", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("received bad status from jumpbox. err: ", resp.Status)
	}

	client := &RemoteConnection{
		sessionRequestChan: sessReqChan,
		conn:               conn,
		out:                gob.NewEncoder(conn),
		in:                 gob.NewDecoder(conn),
		//initialize connection channels
		connectedToCardChan:      make(chan bool, 1),
		identifiedWithServerChan: make(chan bool, 1),
		//initialize card pairing channels
		remoteCertificateChan:    make(chan cert.CardCertificate, 1),
		remoteIdentityChan:       make(chan []byte, 1),
		cardPair1DataChan:        make(chan []byte, 1),
		finalizeCardPairDataChan: make(chan []byte, 1),

		phononAckChan: make(chan bool, 1),

		pairingStatus: model.StatusConnectedToBridge,
	}

	//First send the client cert to kick off connection validation
	if client.localCertificate == nil {
		client.localCertificate, err = client.getLocalCertificate()
		if err != nil {
			log.Error("could not fetch certificate from card: ", err)
			return nil, err
		}
	}
	log.Debug("client has crt: ", client.localCertificate)
	msg := v1.Message{
		Name:    v1.ResponseCertificate,
		Payload: client.localCertificate.Serialize(),
	}
	err = client.out.Encode(msg)
	if err != nil {
		log.Error("unable to send cert to jump server. err: ", err)
		return nil, err
	}

	go client.HandleIncoming()

	select {
	case <-client.identifiedWithServerChan:
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("Verification with server timed out")
	}

	client.pairingStatus = model.StatusConnectedToBridge
	return client, nil
}

func (c *RemoteConnection) HandleIncoming() {
	var err error
	message := v1.Message{}
	err = c.in.Decode(&message)
	for err == nil {
		c.process(message)
		message = v1.Message{}
		err = c.in.Decode(&message)
	}
	log.Printf("Error decoding message: %s", err.Error())
	c.pairingStatus = model.StatusUnconnected
}

func (c *RemoteConnection) process(msg v1.Message) {
	log.Debug(fmt.Sprintf("Processing %s message", msg.Name))
	switch msg.Name {
	case v1.RequestCertificate:
		c.sendCertificate(msg)
	case v1.ResponseCertificate:
		c.receiveCertificate(msg)
	case v1.RequestIdentify:
		c.sendIdentify(msg)
	case v1.ResponseIdentify:
		c.processIdentify(msg)
	case v1.MessageError:
		log.Error(string(msg.Payload))
	case v1.MessageIdentifiedWithServer:
		c.identifiedWithServerChan <- true
		c.identifiedWithServer = true
	case v1.MessageConnectedToCard:
		c.connectedToCardChan <- true
		c.pairingStatus = model.StatusConnectedToCard
	// Card pairing requests and responses
	case v1.RequestCardPair1:
		c.processCardPair1(msg)
	case v1.ResponseCardPair1:
		c.cardPair1DataChan <- msg.Payload
	case v1.RequestFinalizeCardPair:
		c.processFinalizeCardPair(msg)
	case v1.ResponseFinalizeCardPair:
		c.finalizeCardPairDataChan <- msg.Payload
	case v1.MessagePhononAck:
		c.phononAckChan <- true
	case v1.RequestReceivePhonon:
		c.processReceivePhonons(msg)
	case v1.RequestVerifyPaired:
		c.processRequestVerifyPaired(msg)
	case v1.MessageDisconnected:
		c.disconnect()
	case v1.RequestDisconnectFromCard:
		c.disconnectFromCard()
	case v1.ResponseVerifyPaired:
		if c.verifyPairedChan != nil {
			c.verifyPairedChan <- string(msg.Payload)
		}
	}
}

/////
// Below are the request processing methods
/////

func (c *RemoteConnection) sendCertificate(msg v1.Message) {
	c.sendMessage(v1.ResponseCertificate, c.localCertificate.Serialize())
}

func (c *RemoteConnection) sendIdentify(msg v1.Message) {
	_, sig, err := c.requestIdentifyCard(msg.Payload)
	if err != nil {
		log.Error("Issue identifying local card", err.Error())
		return
	}
	payload := []byte{}
	buf := bytes.NewBuffer(payload)
	enc := gob.NewEncoder(buf)
	enc.Encode(sig)
	c.sendMessage(v1.ResponseIdentify, buf.Bytes())
}

func (c *RemoteConnection) processIdentify(msg v1.Message) {
	key, sig, err := card.ParseIdentifyCardResponse(msg.Payload)
	if err != nil {
		log.Error("Issue parsing identify card response", err.Error())
		return
	}
	if !ecdsa.Verify(key, c.counterpartyNonce[:], sig.R, sig.S) {
		log.Error("Unable to verify card challenge")
		return
	} else {
		c.verified = true
		return
	}
}

func (c *RemoteConnection) processCardPair1(msg v1.Message) {
	if c.pairingStatus != model.StatusConnectedToCard {
		log.Error("Card either not connected to a card or already paired")
		return
	}
	cardPairData, err := c.requestCardPair1(msg.Payload)
	if err != nil {
		log.Error("error with card pair 1", err.Error())
		return
	}
	c.pairingStatus = model.StatusCardPair1Complete
	c.sendMessage(v1.ResponseCardPair1, cardPairData)

}

func (c *RemoteConnection) processFinalizeCardPair(msg v1.Message) {
	if c.pairingStatus != model.StatusCardPair1Complete {
		log.Error("Unable to pair. Step one not complete")
		return
	}
	err := c.requestFinalizeCardPair(msg.Payload)
	if err != nil {
		log.Error("Error finalizing Card Pair", err.Error())
		c.sendMessage(v1.ResponseFinalizeCardPair, []byte(err.Error()))
		return
	}
	c.sendMessage(v1.ResponseFinalizeCardPair, []byte{})
	c.pairingStatus = model.StatusPaired
	c.requestSetRemoteCard(c)
	//c.finalizeCardPairErrorChan <- err
}

func (c *RemoteConnection) processReceivePhonons(msg v1.Message) {
	// would check for status to be paired, but for replayability, I'm not entirely sure this is necessary
	err := c.requestReceivePhonons(msg.Payload)
	if err != nil {
		log.Error(err.Error())
		return
	}
	c.sendMessage(v1.MessagePhononAck, []byte{})
}

// ProcessProvideCertificate is for adding a remote card's certificate to the remote portion of the struct
func (c *RemoteConnection) receiveCertificate(msg v1.Message) {
	remoteCert, err := cert.ParseRawCardCertificate(msg.Payload)
	if err != nil {
		log.Error(err)
		return
	}
	c.remoteCertificateChan <- remoteCert
	c.remoteCertificate = &remoteCert
}

/////
// Below are the methods that satisfy the interface for remote counterparty
/////
func (c *RemoteConnection) Identify() error {
	var nonce [32]byte
	rand.Read(nonce[:])
	c.counterpartyNonce = nonce
	c.sendMessage(v1.RequestIdentify, nonce[:])
	select {
	case <-c.remoteIdentityChan:
		return nil
	case <-time.After(10 * time.Second):
		return ErrTimeout

	}
}

func (c *RemoteConnection) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	log.Debug("card pair initiated")
	c.sendMessage(v1.RequestCardPair1, initPairingData)
	select {
	case cardPairData := <-c.cardPair1DataChan:
		return cardPairData, nil
	case <-time.After(10 * time.Second):
		return []byte{}, ErrTimeout
	}
}

func (c *RemoteConnection) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	//unneeded
	return []byte{}, nil
}

func (c *RemoteConnection) FinalizeCardPair(cardPair2Data []byte) error {
	c.sendMessage(v1.RequestFinalizeCardPair, cardPair2Data)
	if !(c.pairingStatus == model.StatusPaired) {
		select {
		case errorbytes := <-c.finalizeCardPairDataChan:
			var err error
			if len(errorbytes) > 0 {
				return errors.New(string(errorbytes))
			} else {
				return err
			}
		case <-time.After(10 * time.Second):
			return ErrTimeout
		}
	}
	c.pairingStatus = model.StatusPaired
	c.requestSetRemoteCard(c)
	return nil
}

func (c *RemoteConnection) GetCertificate() (*cert.CardCertificate, error) {
	if c.remoteCertificate == nil {
		c.sendMessage(v1.RequestCertificate, []byte{})
		select {
		case cert := <-c.remoteCertificateChan:
			c.remoteCertificate = &cert
		case <-time.After(10 * time.Second):
			return nil, ErrTimeout
		}

	}
	return c.remoteCertificate, nil
}

func (c *RemoteConnection) ConnectToCard(cardID string) error {
	log.Info("sending requestConnectCard2Card message")
	c.sendMessage(v1.RequestConnectCard2Card, []byte(cardID))
	var err error
	select {
	case <-time.After(10 * time.Second):
		log.Error("Connection Timed out Waiting for peer")
		c.conn.Close()
		err = ErrTimeout
	case <-c.connectedToCardChan:
		c.pairingStatus = model.StatusConnectedToCard
		err = nil
	}
	cert, err := c.GetCertificate()
	if err != nil {
		return err
	}
	c.remoteCertificate = cert
	return nil
}

func (c *RemoteConnection) ReceivePhonons(PhononTransfer []byte) error {
	c.sendMessage(v1.RequestReceivePhonon, PhononTransfer)
	select {
	case <-time.After(10 * time.Second):
		log.Error("unable to verify remote recipt of phonons")
		return ErrTimeout
	case <-c.phononAckChan:
		return nil
	}
}

func (c *RemoteConnection) GenerateInvoice() (invoiceData []byte, err error) {
	// todo:
	return
}

func (c *RemoteConnection) ReceiveInvoice(invoiceData []byte) error {
	// todo:
	return nil
}

// Utility functions
func (c *RemoteConnection) sendMessage(messageName string, messagePayload []byte) {
	log.Debug(messageName, string(messagePayload))

	tosend := &v1.Message{
		Name:    messageName,
		Payload: messagePayload,
	}
	c.out.Encode(tosend)
}

func (c *RemoteConnection) VerifyPaired() error {
	tosend := &v1.Message{
		Name:    v1.RequestVerifyPaired,
		Payload: []byte(""),
	}
	c.verifyPairedChan = make(chan string)
	c.out.Encode(tosend)

	var connectedCardID string
	select {
	case connectedCardID = <-c.verifyPairedChan:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("counterparty card not paired to this card")
	}
	c.verifyPairedChan = nil

	var err error
	connectedID, err := c.requestGetName()
	if err != nil {
		return err
	}
	if connectedCardID != connectedID {
		//remote isn't paired to this card
		err = c.requestPairWithRemote(c)
	}
	return err
}

func (c *RemoteConnection) processRequestVerifyPaired(msg v1.Message) {
	tosend := &v1.Message{
		Name: v1.ResponseVerifyPaired,
	}
	if c.pairingStatus == model.StatusPaired {
		key, err := util.ParseECCPubKey(c.remoteCertificate.PubKey)
		if err != nil {
			//oopsie
			return
		}
		msg := util.CardIDFromPubKey(key)
		tosend.Payload = []byte(msg)
	}
	c.out.Encode(tosend)
}

func (c *RemoteConnection) PairingStatus() model.RemotePairingStatus {
	return c.pairingStatus
}

func (c *RemoteConnection) disconnect() {
	c.pairingStatus = model.StatusUnconnected
	c.requestSetPaired(false)
}

func (c *RemoteConnection) disconnectFromCard() {
	c.pairingStatus = model.StatusConnectedToBridge
	c.requestSetPaired(false)
}
