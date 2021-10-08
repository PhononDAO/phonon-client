package remote

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/posener/h2conn"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

type remoteCounterParty struct {
	conn *h2conn.Conn
}

func Connect(url string, ignoreTLS bool) (*remoteCounterParty, error) {
	d := &h2conn.Client{
		Client: &http.Client{
			Transport: &http2.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreTLS}},
		},
	}
	conn, _, err := d.Connect(context.Background(), url) //url)
	if err != nil {
		return &remoteCounterParty{}, fmt.Errorf("Unable to connect to remote server %e,", err)
	}
	counterParty := &remoteCounterParty{
		conn: conn,
	}
	go counterParty.HandleIncoming()

	return counterParty, nil
}

// memory leak ohh boy!
func (c *remoteCounterParty) HandleIncoming() {
	cmdDecoder := json.NewDecoder(c.conn)
	messageChan := make(chan (Request))

	go func(msgchan chan Request) {
		defer close(msgchan)
		for {
			message := Request{}
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

func (c *remoteCounterParty) process(req Request) {
	fmt.Printf("%+v", req)
}

func (c *remoteCounterParty) GetCertificate() (cert.CardCertificate, error) {
	// generate Request to server
	return cert.CardCertificate{}, nil
}

func (c *remoteCounterParty) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	// generate request
	// add initPairingData to request
	// send it off
	return
}

func (c *remoteCounterParty) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error) {
	// generate request
	// add cardPairData to request
	// send it off
	return
}

func (c *remoteCounterParty) FinalizeCardPair(cardPair2Data []byte) error {
	// generate request
	// add cardPair2Data to request
	// send it off
	return nil
}

func (c *remoteCounterParty) ReceivePhonons(PhononTransfer []byte) error {

	return nil
}

func (c *remoteCounterParty) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error) {
	// todo: figure this one out
	return
}

func (c *remoteCounterParty) GenerateInvoice() (invoiceData []byte, err error) {
	// todo: uhhhhhhh
	return
}

func (c *remoteCounterParty) ReceiveInvoice(invoiceData []byte) error {
	// todo: oh boy
	return nil
}
