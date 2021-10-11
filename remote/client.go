package remote

import (
	"context"
	"crypto/tls"
	"encoding/gob"
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

type remoteConnection struct {
	conn              *h2conn.Conn
	encoder           *gob.Encoder
	remoteCertificate *cert.CardCertificate
	session           *card.Session
}

func Connect(url string, ignoreTLS bool) (*remoteConnection, error) {
	d := &h2conn.Client{
		Client: &http.Client{
			Transport: &http2.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreTLS}},
		},
	}
	conn, _, err := d.Connect(context.Background(), url) //url)
	if err != nil {
		return &remoteConnection{}, fmt.Errorf("Unable to connect to remote server %e,", err)
	}
	remoteConn := &remoteConnection{
		conn: conn,
	}
	go remoteConn.HandleIncoming()

	return remoteConn, nil
}

// memory leak ohh boy!
func (c *remoteConnection) HandleIncoming() {
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

func (c *remoteConnection) process(msg Message) {
	switch msg.Name {
	case RequestProvideCertifcate:
		c.sendCertificate(msg)
	case ResponseProvideCertificate:
		c.ProcessProvideCertificate(msg)
	}
}

// Below are the request processing methods
func (c *remoteConnection) sendCertificate(msg Message) {
	certbytes := c.session.Cert.Serialize()
	resp := Message{
		Name:    ResponseProvideCertificate,
		Payload: certbytes,
	}
	c.encoder.Encode(resp)
}

// ProcessProvideCertificate is for adding a remote card's certificate to the remote portion of the struct
func (c *remoteConnection) ProcessProvideCertificate(msg Message) {
	remoteCert, err := cert.ParseRawCardCertificate(msg.Payload)
	if err != nil {
		//handle this
	}
	c.remoteCertificate = &remoteCert
}

// Below are the methods that satisfy the interface for remote counterparty

func (c *remoteConnection) GetCertificate() (cert.CardCertificate, error) {
	toSend := Message{
		Name: RequestProvideCertifcate,
	}
	c.encoder.Encode(toSend)
	for c.remoteCertificate == nil {
		time.Sleep(time.Second * 5)
	}
	return *c.remoteCertificate, nil
}

func (c *remoteConnection) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	// generate request
	// add initPairingData to request
	// send it off
	return
}

func (c *remoteConnection) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	// generate request
	// add cardPairData to request
	// send it off
	return
}

func (c *remoteConnection) FinalizeCardPair(cardPair2Data []byte) error {
	// generate request
	// add cardPair2Data to request
	// send it off
	return nil
}

func (c *remoteConnection) ReceivePhonons(PhononTransfer []byte) error {

	return nil
}

func (c *remoteConnection) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error) {
	// todo: figure this one out
	return
}

func (c *remoteConnection) GenerateInvoice() (invoiceData []byte, err error) {
	// todo: uhhhhhhh
	return
}

func (c *remoteConnection) ReceiveInvoice(invoiceData []byte) error {
	// todo: oh boy
	return nil
}
