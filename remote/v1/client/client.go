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
	logger                   *log.Entry

	phononAckChan chan bool
}

var ErrTimeout = errors.New("Timeout")

// Requests into the card session
func (c *RemoteConnection) getLocalCertificate() (*cert.CardCertificate, error) {
	req := &model.RequestCertificate{
		Ret: make(chan model.ResponseCertificate),
	}
	c.logger.Debug("Requesting local card certificate")
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
	c.logger.Debug("Requesting Identify card")
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.PubKey, ret.Sig, ret.Err
}

func (c *RemoteConnection) requestCardPair1(payload []byte) ([]byte, error) {
	req := &model.RequestCardPair1{
		Ret:     make(chan model.ResponseCardPair1),
		Payload: payload,
	}
	c.logger.Debug("Requesting card pair 1")
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Payload, ret.Err
}

func (c *RemoteConnection) requestFinalizeCardPair(payload []byte) error {
	req := &model.RequestFinalizeCardPair{
		Ret:     make(chan model.ResponseFinalizeCardPair),
		Payload: payload,
	}
	c.logger.Debug("Requesting finalize card pair")
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func (c *RemoteConnection) requestReceivePhonons(payload []byte) error {
	req := &model.RequestReceivePhonons{
		Ret:     make(chan model.ResponseReceivePhonons),
		Payload: payload,
	}
	c.logger.Debug("Requesting Receive Phonons")
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Err
}

func (c *RemoteConnection) requestGetName() (string, error) {
	req := &model.RequestGetName{
		Ret: make(chan model.ResponseGetName),
	}
	c.logger.Debug("Requesting Name")
	c.sessionRequestChan <- req
	ret := <-req.Ret
	return ret.Name, ret.Err
}

func (c *RemoteConnection) requestPairWithRemote(card model.CounterpartyPhononCard) error {
	req := &model.RequestPairWithRemote{
		Ret:  make(chan model.ResponsePairWithRemote),
		Card: card,
	}
	c.logger.Debug("Requesting pairing")
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
		return nil, fmt.Errorf("unable to connect to remote server %e,", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("received bad status from jumpbox. err: ", resp.Status)
	}

	client := &RemoteConnection{
		conn:                     conn,
		out:                      gob.NewEncoder(conn),
		in:                       gob.NewDecoder(conn),
		remoteCertificate:        nil,
		localCertificate:         nil,
		sessionRequestChan:       sessReqChan,
		identifiedWithServerChan: make(chan bool, 1),
		identifiedWithServer:     false,
		counterpartyNonce:        [32]byte{},
		verified:                 false,
		connectedToCardChan:      make(chan bool, 1),
		verifyPairedChan:         make(chan string),
		remoteCertificateChan:    make(chan cert.CardCertificate, 1),
		remoteIdentityChan:       make(chan []byte, 1),
		cardPair1DataChan:        make(chan []byte, 1),
		finalizeCardPairDataChan: make(chan []byte, 1),
		pairingStatus:            model.StatusUnconnected,
		logger:                   log.WithField("cardID", "unknown"),
		phononAckChan:            make(chan bool, 1),
	}

	name, err := client.requestGetName()
	if err != nil {
		return nil, err
	}
	client.logger = log.WithField("cardID", name)
	//First send the client cert to kick off connection validation
	client.logger.Debugf("certificate: % X", client.localCertificate)
	client.localCertificate, err = client.getLocalCertificate()
	if err != nil {
		client.logger.Error("could not fetch certificate from card: ", err)
		return nil, err
	}
	client.logger.Debug("client has crt: ", client.localCertificate)
	msg := v1.Message{
		Name:    v1.ResponseCertificate,
		Payload: client.localCertificate.Serialize(),
	}
	err = client.out.Encode(msg)
	if err != nil {
		client.logger.Error("unable to send cert to jump server. err: ", err)
		return nil, err
	}

	go client.HandleIncoming()

	select {
	case <-client.identifiedWithServerChan:
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("verification with server timed out")
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
	c.logger.Printf("Error decoding message: %s", err.Error())
	c.pairingStatus = model.StatusUnconnected
}

func (c *RemoteConnection) process(msg v1.Message) {
	c.logger.Debug(fmt.Sprintf("processing %s message", msg.Name))
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
		c.logger.Error(string(msg.Payload))
	case v1.MessageIdentifiedWithServer:
		c.identifiedWithServerChan <- true
		c.identifiedWithServer = true
	case v1.MessageConnectedToCard:
		c.processConnectedToCard(msg)
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

func (c *RemoteConnection) processConnectedToCard(msg v1.Message) {
	log.Debug("Processing connected to card message")
	counterpartyCert, err := cert.ParseRawCardCertificate(msg.Payload)
	if err != nil {
		c.logger.Error("unable to process counterparty card certificate: ", err.Error())
		return
	}
	c.remoteCertificate = &counterpartyCert
	c.connectedToCardChan <- true
	c.pairingStatus = model.StatusConnectedToCard

}

func (c *RemoteConnection) sendCertificate(msg v1.Message) {
	c.logger.Debug("caching counterparty certificate")
	c.sendMessage(v1.ResponseCertificate, c.localCertificate.Serialize())
}

func (c *RemoteConnection) sendIdentify(msg v1.Message) {
	_, sig, err := c.requestIdentifyCard(msg.Payload)
	if err != nil {
		c.logger.Error("Issue identifying local card", err.Error())
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
		c.logger.Error("Issue parsing identify card response", err.Error())
		return
	}
	if !ecdsa.Verify(key, c.counterpartyNonce[:], sig.R, sig.S) {
		c.logger.Error("Unable to verify card challenge")
		return
	} else {
		c.verified = true
		return
	}
}

func (c *RemoteConnection) processCardPair1(msg v1.Message) {
	if c.pairingStatus != model.StatusConnectedToCard {
		c.logger.Error("Card either not connected to a card or already paired")
		return
	}
	cardPairData, err := c.requestCardPair1(msg.Payload)
	if err != nil {
		c.logger.Error("error with card pair 1", err.Error())
		return
	}
	c.pairingStatus = model.StatusCardPair1Complete
	c.sendMessage(v1.ResponseCardPair1, cardPairData)

}

func (c *RemoteConnection) processFinalizeCardPair(msg v1.Message) {
	if c.pairingStatus != model.StatusCardPair1Complete {
		c.logger.Error("Unable to pair. Step one not complete")
		return
	}
	err := c.requestFinalizeCardPair(msg.Payload)
	if err != nil {
		c.logger.Error("Error finalizing Card Pair", err.Error())
		c.sendMessage(v1.ResponseFinalizeCardPair, []byte(err.Error()))
		return
	}
	c.sendMessage(v1.ResponseFinalizeCardPair, []byte{})
	c.pairingStatus = model.StatusPaired
	//c.finalizeCardPairErrorChan <- err
}

func (c *RemoteConnection) processReceivePhonons(msg v1.Message) {
	// would check for status to be paired, but for replayability, I'm not entirely sure this is necessary
	err := c.requestReceivePhonons(msg.Payload)
	if err != nil {
		c.logger.Error(err.Error())
		return
	}
	c.sendMessage(v1.MessagePhononAck, []byte{})
}

// ProcessProvideCertificate is for adding a remote card's certificate to the remote portion of the struct
func (c *RemoteConnection) receiveCertificate(msg v1.Message) {
	remoteCert, err := cert.ParseRawCardCertificate(msg.Payload)
	if err != nil {
		c.logger.Error(err)
		return
	}
	c.logger.Debug("Remote Certificate received")
	c.remoteCertificateChan <- remoteCert
}

// ///
// Below are the methods that satisfy the interface for remote counterparty
// ///
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
	c.logger.Debug("card pair initiated")
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
	return nil
}

func (c *RemoteConnection) GetCertificate() (*cert.CardCertificate, error) {
	if c.remoteCertificate == nil {
		c.logger.Debug("remote certificate not cached, requesting it")
		c.sendMessage(v1.RequestCertificate, []byte{})
		select {
		case cert := <-c.remoteCertificateChan:
			c.remoteCertificate = &cert
		case <-time.After(10 * time.Second):
			c.logger.Debug("Certificate request timed out")
			return nil, ErrTimeout
		}

	} else {
		c.logger.Debugf("returning cached remote certificate: % X", c.remoteCertificate.Serialize())
	}
	return c.remoteCertificate, nil
}

func (c *RemoteConnection) ConnectToCard(cardID string) error {
	c.logger.Info("sending requestConnectCard2Card message")
	c.sendMessage(v1.RequestConnectCard2Card, []byte(cardID))
	var err error
	select {
	case <-time.After(10 * time.Second):
		c.logger.Error("Connection Timed out Waiting for peer")
		c.conn.Close()
		err = ErrTimeout
		return err
	case <-c.connectedToCardChan:
		c.pairingStatus = model.StatusConnectedToCard
		err = nil
	}
	_, err = c.GetCertificate()
	if err != nil {
		return err
	}
	return nil
}

func (c *RemoteConnection) ReceivePhonons(PhononTransfer []byte) error {
	c.sendMessage(v1.RequestReceivePhonon, PhononTransfer)
	select {
	case <-time.After(10 * time.Second):
		c.logger.Error("unable to verify remote recipt of phonons")
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
	c.logger.Debug(messageName, string(messagePayload))

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
		return fmt.Errorf("counterparty card not paired to this card: timeout")
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
	c.logger.Debug("Pairing Verified")
	return err
}

func (c *RemoteConnection) processRequestVerifyPaired(msg v1.Message) {
	tosend := &v1.Message{
		Name: v1.ResponseVerifyPaired,
	}
	if c.pairingStatus == model.StatusPaired {
		if c.remoteCertificate == nil || c.remoteCertificate.PubKey == nil {
			c.logger.Error("Remote certificate not cached")
			return
		}
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
}

func (c *RemoteConnection) disconnectFromCard() {
	c.pairingStatus = model.StatusConnectedToBridge
}
