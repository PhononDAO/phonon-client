package remote

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	conn, resp, err := d.Connect(context.TODO(), url)
	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return &remoteCounterParty{}, fmt.Errorf("Unable to connect to remote server %e, %s", err, string(body))
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
			err := cmdDecoder.Decode(&message)
			if err != nil {
				log.Info("Error receiving message from connected client")
				return
			}
			msgchan <- message
		}
	}(messageChan)
	
	for message := range messageChan {
		c.process(message)
	}
}

func(c *remoteCounterParty)process(Request){
	//todo
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
