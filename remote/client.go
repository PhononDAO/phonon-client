package remote

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
	"github.com/posener/h2conn"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

type RemoteConnection struct {
	conn                      *h2conn.Conn
	encoder                   *gob.Encoder
	remoteCertificate         *cert.CardCertificate
	session                   *card.Session
	remoteCertificateChan     chan cert.CardCertificate
	cardPairDataChan          chan []byte
	cardPairData2Chan         chan []byte
	remoteIdentityChan        chan []byte
	identifiedWithServerChan  chan bool
	finalizeCardPairErrorChan chan error
	identifiedWithServer      bool
	counterpartyNonce         [32]byte
	verified                  bool
	connectedToCardChan       chan bool
}

// this will go someplace, I swear
var ErrTimeout = errors.New("Timeout")

func Connect(s *card.Session, url string, ignoreTLS bool) (*RemoteConnection, error) {
	d := &h2conn.Client{
		Client: &http.Client{
			Transport: &http2.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreTLS}},
		},
	}

	conn, _, err := d.Connect(context.Background(), url) //url)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to remote server %e,", err)
	}
	remoteConn := &RemoteConnection{
		conn:                      conn,
		remoteCertificateChan:     make(chan cert.CardCertificate, 1),
		cardPairDataChan:          make(chan []byte, 1),
		cardPairData2Chan:         make(chan []byte, 1),
		remoteIdentityChan:        make(chan []byte, 1),
		connectedToCardChan:       make(chan bool, 1),
		identifiedWithServerChan:  make(chan bool, 1),
		finalizeCardPairErrorChan: make(chan error, 1),
	}

	go remoteConn.HandleIncoming()
	remoteConn.encoder = gob.NewEncoder(conn)
	remoteConn.session = s
	return remoteConn, nil
}

// memory leak ohh boy!
func (c *RemoteConnection) HandleIncoming() {
	cmdDecoder := gob.NewDecoder(c.conn)
	messageChan := make(chan (Message))

	go func(msgchan chan Message) {
		defer close(msgchan)
		for {
			message := Message{}
			//todo read raw and decode separately to avoid killing the whole thing on a malformed message
			err := cmdDecoder.Decode(&message)
			if err != nil {
				log.Info("Error receiving message from connected server")
				return
			}
			msgchan <- message
		}
	}(messageChan)

	for message := range messageChan {
		c.process(message)
	}
}

func (c *RemoteConnection) process(msg Message) {
	fmt.Printf("processing message: %s\nPayload: %+v\nPayloadString: %s\n", msg.Name, msg.Payload, string(msg.Payload))
	switch msg.Name {
	case RequestCertificate:
		c.sendCertificate(msg)
	case ResponseCertificate:
		c.receiveCertificate(msg)
	case RequestIdentify:
		c.sendIdentify(msg)
	case ResponseIdentify:
		c.ProcessIdentify(msg)
	case RequestCardPair1:
		c.ProcessCardPair1(msg)
	case RequestCardPair2:
		c.ProcessCardPair2(msg)
	case RequestFinalizeCardPair:
		c.ProcessFinalizeCardPair(msg)
	case MessageError:
		fmt.Println(string(msg.Payload))
	case MessageIdentifiedWithServer:
		c.identifiedWithServerChan <- true
		c.identifiedWithServer = true
	case MessageConnectedToCard:
		c.connectedToCardChan <- true
	}
}

/////
// Below are the request processing methods
/////

func (c *RemoteConnection) sendCertificate(msg Message) {
	fmt.Println(c.session.Cert)
	cert, err := c.session.GetCertificate()
	if err != nil {
		log.Error("Cert doesn't exist")
	}
	c.sendMessage(ResponseCertificate, cert.Serialize())
}

func (c *RemoteConnection) sendIdentify(msg Message) {
	fmt.Println(msg.Payload)
	_, sig, err := c.session.IdentifyCard(msg.Payload)
	if err != nil {
		log.Error("Issue identifying local card", err.Error())
		return
	}
	payload := []byte{}
	buf := bytes.NewBuffer(payload)
	enc := gob.NewEncoder(buf)
	enc.Encode(sig)
	c.sendMessage(ResponseIdentify, buf.Bytes())
}

func (c *RemoteConnection) ProcessIdentify(msg Message) {
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

func (c *RemoteConnection) ProcessCardPair1(msg Message) {
	cardPairData, err := c.session.CardPair(msg.Payload)
	if err != nil {
		log.Error("error with card pair 1", err.Error())
	}
	c.sendMessage(RequestCardPair2, cardPairData)

}

func (c *RemoteConnection) ProcessCardPair2(msg Message) {
	// handle this error
	cardPair2Data, err := c.session.CardPair2(msg.Payload)
	if err != nil {
		log.Error("Error with Card pair 2", err.Error())
	}
	c.sendMessage(RequestFinalizeCardPair, cardPair2Data)
}

func (c *RemoteConnection) ProcessFinalizeCardPair(msg Message) {
	err := c.session.FinalizeCardPair(msg.Payload)
	if err != nil {
		log.Error("Error finalizing Card Pair", err.Error())
	}
	c.finalizeCardPairErrorChan<- err
}

// ProcessProvideCertificate is for adding a remote card's certificate to the remote portion of the struct
func (c *RemoteConnection) receiveCertificate(msg Message) {
	remoteCert, err := cert.ParseRawCardCertificate(msg.Payload)
	if err != nil {
		fmt.Println(err)
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
	c.sendMessage(RequestIdentify, nonce[:])
	select {
	case <-c.remoteIdentityChan:
		return nil
	case <-time.After(10 * time.Second):
		return ErrTimeout

	}
}

func (c *RemoteConnection) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	c.sendMessage(RequestCardPair1, initPairingData)
	select {
	case cardPairData := <-c.cardPairDataChan:
		return cardPairData, nil
	case <-time.After(10 * time.Second):
		return []byte{}, ErrTimeout

	}
}

func (c *RemoteConnection) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	c.sendMessage(RequestCardPair2, cardPairData)
	select {
	case cardPairData2 := <-c.cardPairData2Chan:
		return cardPairData2, nil
	case <-time.After(10 * time.Second):
		return []byte{}, ErrTimeout
	}
}

func (c *RemoteConnection) FinalizeCardPair(cardPair2Data []byte) error {
	c.sendMessage(RequestFinalizeCardPair, cardPair2Data)
	select {
	case err := <-c.finalizeCardPairErrorChan:
		return err
	case <-time.After(10 * time.Second):
		return ErrTimeout
	}
}

func (c *RemoteConnection) GetCertificate() (*cert.CardCertificate, error) {
	if c.remoteCertificate == nil {
		c.sendMessage(RequestCertificate, []byte{})
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
	fmt.Println("CARDID:", cardID)
	if !c.identifiedWithServer {
		select {
		case <-time.After(10 * time.Second):
			return ErrTimeout
		case <-c.identifiedWithServerChan:
			fmt.Println("received Identified with server")
		}
	}
	fmt.Println("sending requestConnectCard2Card message")
	c.sendMessage(RequestConnectCard2Card, []byte(cardID))
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("Connection Timed out Waiting for peer")
		c.conn.Close()
		return ErrTimeout
	case <-c.connectedToCardChan:
		return nil
	}
}

func (c *RemoteConnection) ReceivePhonons(PhononTransfer []byte) error {
	//	PhononTransfer <- c.receivePhononChan
	return nil
}

func (c *RemoteConnection) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error) {
	// todo: figure this one out
	return
}

func (c *RemoteConnection) GenerateInvoice() (invoiceData []byte, err error) {
	// todo: uhhhhhhh
	return
}

func (c *RemoteConnection) ReceiveInvoice(invoiceData []byte) error {
	// todo: oh boy
	return nil
}

// Utility functions
func (c *RemoteConnection) sendMessage(messageName string, messagePayload []byte) {
	fmt.Println(messageName, string(messagePayload))

	tosend := &Message{
		Name:    messageName,
		Payload: messagePayload,
	}

	c.encoder.Encode(tosend)
}
