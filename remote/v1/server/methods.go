package server

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"reflect"
	"strings"

	"github.com/GridPlus/phonon-client/cert"
	v1 "github.com/GridPlus/phonon-client/remote/v1"
	"github.com/GridPlus/phonon-client/util"
	log "github.com/sirupsen/logrus"
)

func (c *clientSession) RequestIdentify() (challengeNonce []byte, err error) {
	challengeNonce = make([]byte, 32)
	_, err = rand.Reader.Read(challengeNonce)
	if err != nil {
		log.Error("unable to generate challenge nonce. err: ", err)
		return nil, err
	}
	err = c.out.Encode(v1.Message{Name: v1.RequestIdentify, Payload: challengeNonce})
	if err != nil {
		log.Error("unable to send identify request")
		return nil, err
	}
	return challengeNonce, nil
}

func (c *clientSession) ReceiveIdentifyResponse() (*util.ECDSASignature, error) {
	var identifyResp v1.Message
	var sig util.ECDSASignature
	err := c.in.Decode(&identifyResp)
	if err != nil {
		log.Error("could not receive identify response. err: ", err)
		return nil, err
	}
	log.Infof("received identify response: %+v\n", identifyResp)
	if identifyResp.Name == v1.ResponseIdentify {
		buf := bytes.NewBuffer(identifyResp.Payload)
		decoder := gob.NewDecoder(buf)
		err := decoder.Decode(&sig)
		if err != nil {
			log.Error("unable to decode sig. err: ", err)
			return nil, err
		}
	}
	log.Info("returning sig")
	return &sig, nil
}

func (c *clientSession) provideCertificate() {
	if c.Counterparty == nil {
		c.out.Encode(v1.Message{
			Name:    v1.MessageError,
			Payload: []byte("no counterparty connected. Cannot get certificate"),
		})
		return
	}
	if reflect.DeepEqual(c.Counterparty.certificate, cert.CardCertificate{}) {
		c.out.Encode(v1.Message{
			Name:    v1.MessageError,
			Payload: []byte("failed to retrieve cached counterparty certificate"),
		})
		return

	}
	msg := v1.Message{
		Name:    v1.ResponseCertificate,
		Payload: c.Counterparty.certificate.Serialize(),
	}
	err := c.out.Encode(msg)
	if err != nil {
		log.Error("error encoding provideCertificate reply: ", err)
		return
	}
}

//Start of alternate implementation using pairing map
// func (c *clientSession) ConnectCard2Card(counterpartyID string) {
// 	for {
// 		if counterparty, ok := clientSessions[counterpartyID]; ok {
// 			log.Info("counterparty found, connecting %v to %v", c.Name, counterparty)
// 			//generate hash representing pairing
// 			//TODO: make this more bulletproof, collisions are semi possible
// 			var pairingData []byte
// 			var p pairing
// 			if c.Name < counterparty.Name {
// 				pairingData = append([]byte(c.Name), []byte(counterparty.Name)...)
// 				p = pairing{
// 					initiator: c,
// 					responder: counterparty,
// 				}
// 			} else {
// 				pairingData = append([]byte(counterparty.Name), []byte(c.Name)...)
// 				p = pairing{
// 					initiator: counterparty,
// 					responder: c,
// 				}
// 			}
// 			pairingHash := sha256.Sum256(pairingData)
// 			pairingID := string(pairingHash[:])
// 			pairings[pairingID] = p

// 		}
// 		time.Sleep(250 * time.Millisecond)
// 	}

// }

func (c *clientSession) ConnectCard2Card(msg v1.Message) {
	log.Infof("attempting to connect card %s to card %s\n", c.Name, string(msg.Payload))
	counterparty, ok := clientSessions[strings.ToLower(string(msg.Payload))]
	if !ok {
		c.out.Encode(v1.Message{
			Name:    v1.MessageError,
			Payload: []byte("No connected card"),
		})
		log.Error("no connected session:", string(msg.Payload))
		return
	} else if counterparty.Counterparty == nil && c.Counterparty == nil {
		counterparty.Counterparty = c
		c.Counterparty = counterparty
		c.out.Encode(v1.Message{
			Name:    v1.MessageConnectedToCard,
			Payload: c.Counterparty.certificate.Serialize(),
		})
		c.Counterparty.out.Encode(v1.Message{
			Name:    v1.MessageConnectedToCard,
			Payload: c.certificate.Serialize(),
		})
		log.Infof("Connected card %s to card %s\n", c.Name, c.Counterparty.Name)
	} else if c.Counterparty == counterparty && counterparty.Counterparty == c {
		//do nothing
	} else {
		c.out.Encode(v1.Message{
			Name:    v1.MessageError,
			Payload: []byte("Unable to connect. Connection already satisfied"),
		})
	}
}

func (c *clientSession) disconnectFromCard(msg v1.Message) {
	out := v1.Message{
		Name: v1.RequestDisconnectFromCard,
	}
	// encode can fail, so it needs to be checked. Not sure how to handle that
	if c.Counterparty != nil && c.Counterparty.out != nil {
		c.Counterparty.out.Encode(out)
	}
	if c.out != nil {
		c.out.Encode(out)
	}
	if c.Counterparty != nil {
		c.Counterparty.Counterparty = nil
	}
	c.Counterparty = nil
}

func (c *clientSession) endSession(msg v1.Message) {
	c.disconnectFromCard(msg)
	delete(clientSessions, c.Name)
	if c.underlyingConn != nil {
		c.underlyingConn.Close()
	}
}

func (c *clientSession) noop(msg v1.Message) {
	// don't do anything
	// this is eventually going to be for preventing connection timeouts, but may not be nessesary in the future
}

func (c *clientSession) passthrough(msg v1.Message) {
	if c.Counterparty == nil {
		log.Debug("Passing through message to counterparty")
		ret := v1.Message{
			Name: v1.MessagePassthruFailed,
		}
		c.out.Encode(ret)
	} else {
		c.Counterparty.out.Encode(msg)
	}
}

func (c *clientSession) RequestSendPhonon(msg v1.Message) {

}

func (c *clientSession) RequestPhononAck(msg v1.Message) {

}

func (c *clientSession) sendPhonon(msg v1.Message) {
	// save this packet for later
	// delete after ack
}

func (c *clientSession) ack(msg v1.Message) {
	// delete saved phonon when received
}
