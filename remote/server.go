package remote

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/util"
	"github.com/posener/h2conn"
	log "github.com/sirupsen/logrus"
)

func StartServer(port string, certfile string, keyfile string) {
	//init sessions global
	clientSessions = make(map[string]*clientSession)
	http.HandleFunc("/phonon", handle)
	http.HandleFunc("/connected", listConnected)
	http.HandleFunc("/", index)
	err := http.ListenAndServeTLS(":"+port, certfile, keyfile, nil)
	if err != nil {
		log.Errorf("Error with web server:, %s", err.Error())
	}
}

type clientSession struct {
	Name           string
	certificate    cert.CardCertificate
	underlyingConn *h2conn.Conn
	out            *gob.Encoder
	in             *gob.Decoder
	validated      bool
	Counterparty   *clientSession
	// the same name that goes in the lookup value of the clientSession map
}

var clientSessions map[string]*clientSession

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello there"))
}

func listConnected(w http.ResponseWriter, r *http.Request) {
	ret, _ := json.Marshal(clientSessions)
	w.Write(ret)
}

func handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h2conn.Accept(w, r)
	if err != nil {
		log.Error("Unable to establish http2 duplex connection with ", r.RemoteAddr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	defer conn.Close()

	cmdEncoder := gob.NewEncoder(conn)
	cmdDecoder := gob.NewDecoder(conn)
	//generate session
	session := clientSession{
		Name:           "",
		certificate:    cert.CardCertificate{},
		underlyingConn: conn,
		out:            cmdEncoder,
		in:             cmdDecoder,
		validated:      false,
		Counterparty:   nil,
	}

	valid, err := session.ValidateClient()
	if err != nil {
		err = session.out.Encode(err.Error())
		if err != nil {
			log.Error("failed sending cert validation failure response: ", err)
			return
		}
	}
	if !valid {
		//TODO: use a real error, possibly from cert package
		msg := Message{
			Name:    MessageDisconnected,
			Payload: []byte("Certificate invalid"),
		}
		err = session.out.Encode(msg)
		if err != nil {
			log.Error("failed sending invalid cert response: ", err)
			return
		}
		log.Error("certificate invalid")
	}
	log.Info("validated client connection: ", session.Name)

	//Client is now validated, move on
	for r.Context().Err() == nil {
		var msg Message
		err := session.in.Decode(&msg)
		if err != nil {
			log.Error("failed receiving message: ", err)
			return
		}
		log.Debugf("received %v message with payload: % X\n", msg.Name, msg.Payload)
		err = session.process(msg)
		if err != nil {
			log.Errorf("failed to process incoming %v msg. err: %v", msg.Name, err)
			log.Errorf("msg payload: % X", msg.Payload)
		}
	}
}

func (c *clientSession) process(msg Message) error {
	switch msg.Name {
	case RequestConnectCard2Card:
		c.ConnectCard2Card(msg)
	case RequestDisconnectFromCard:
		c.disconnectFromCard(msg)
	case RequestEndSession:
		c.endSession(msg)
	case RequestNoOp:
		c.noop(msg)
	case RequestIdentify, ResponseIdentify, RequestCardPair1, ResponseCardPair1, RequestCardPair2, ResponseCardPair2, RequestFinalizeCardPair, ResponseFinalizeCardPair, RequestReceivePhonon, MessagePhononAck, RequestVerifyPaired, ResponseVerifyPaired:
		c.passthrough(msg)
	case RequestCertificate:
		c.provideCertificate()
	}
	//TODO: provide actual errors, or ensure all the cases handle errors themselves
	return nil
}

func (c *clientSession) ValidateClient() (bool, error) {
	log.Info("validating client connection")
	//Read client certificate
	var in Message
	err := c.in.Decode(&in)
	if err != nil {
		log.Error("unable to decode raw client certificate bytes: ", err)
		return false, err
	}
	log.Info("past first Decode:")
	c.certificate, err = cert.ParseRawCardCertificate(in.Payload)
	if err != nil {
		log.Infof("failed to parse certificate from client %s\n", err.Error())
		return false, err
	}
	log.Info("parsed cert: ", c.certificate)
	//Validate certificate is signed by valid origin

	//Send Identify Card Challenge
	challengeNonce, err := c.RequestIdentify()
	if err != nil {
		log.Error("failed to send IDENTIFY_CARD request: ", err)
		return false, nil
	}

	sig, err := c.ReceiveIdentifyResponse()
	if err != nil {
		log.Error("failed to receive IDENTIFY_CARD response: ", err)
		return false, err
	}
	log.Infof("received sig from identifyResponse: %+v", sig)
	key, err := util.ParseECCPubKey(c.certificate.PubKey)
	if err != nil {
		log.Error("Unable to parse pubkey from certificate", err.Error())
		return false, err
	}
	if !ecdsa.Verify(key, challengeNonce, sig.R, sig.S) {
		log.Error("unable to verify card challenge")
		return false, err
	}

	//Cert has been validated, register clientSession with server and grab card name
	c.validated = true
	name := util.CardIDFromPubKey(key)
	c.Name = name
	clientSessions[name] = c
	c.out.Encode(Message{
		Name:    MessageIdentifiedWithServer,
		Payload: []byte(name),
	})

	//Return to main loop to process further client requests
	return true, nil
}

func (c *clientSession) RequestIdentify() (challengeNonce []byte, err error) {
	challengeNonce = make([]byte, 32)
	_, err = rand.Reader.Read(challengeNonce)
	if err != nil {
		log.Error("unable to generate challenge nonce. err: ", err)
		return nil, err
	}
	err = c.out.Encode(Message{Name: RequestIdentify, Payload: challengeNonce})
	if err != nil {
		log.Error("unable to send identify request")
		return nil, err
	}
	return challengeNonce, nil
}

func (c *clientSession) ReceiveIdentifyResponse() (*util.ECDSASignature, error) {
	var identifyResp Message
	var sig util.ECDSASignature
	err := c.in.Decode(&identifyResp)
	if err != nil {
		log.Error("could not receive identify response. err: ", err)
		return nil, err
	}
	log.Infof("received identify response: %+v\n", identifyResp)
	if identifyResp.Name == ResponseIdentify {
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
		c.out.Encode(Message{
			Name:    MessageError,
			Payload: []byte("No counterparty connected. Cannot get certificate"),
		})
		return
	}
	msg := Message{
		Name:    ResponseCertificate,
		Payload: c.Counterparty.certificate.Serialize(),
	}
	err := c.out.Encode(msg)
	if err != nil {
		log.Error("Error encoding provideCertificate reply: ", err)
		return
	}
	return
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

func (c *clientSession) ConnectCard2Card(msg Message) {
	log.Infof("attempting to connect card %s to card %s\n", c.Name, string(msg.Payload))
	counterparty, ok := clientSessions[string(msg.Payload)]
	if !ok {
		c.out.Encode(Message{
			Name:    MessageError,
			Payload: []byte("No connected card"),
		})
		log.Error("No connected session:", string(msg.Payload))
		return
	} else if counterparty.Counterparty == nil && c.Counterparty == nil {
		counterparty.Counterparty = c
		c.Counterparty = counterparty
		c.out.Encode(Message{
			Name: MessageConnectedToCard,
		})
		c.Counterparty.out.Encode(Message{
			Name: MessageConnectedToCard,
		})
		log.Infof("Connected card %s to card %s\n", c.Name, c.Counterparty.Name)
	} else if c.Counterparty == counterparty && counterparty.Counterparty == c {
		//do nothing
	} else {
		c.out.Encode(Message{
			Name:    MessageError,
			Payload: []byte("Unable to connect. Connection already satisfied"),
		})
	}
}

func (c *clientSession) disconnectFromCard(msg Message) {
	out := Message{
		Name: RequestDisconnectFromCard,
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

func (c *clientSession) endSession(msg Message) {
	c.disconnectFromCard(msg)
	delete(clientSessions, c.Name)
	if c.underlyingConn != nil {
		c.underlyingConn.Close()
	}
}

func (c *clientSession) noop(msg Message) {
	// don't do anything
	// this is eventually going to be for preventing connection timeouts, but may not be nessesary in the future
}

func (c *clientSession) passthrough(msg Message) {
	if c.Counterparty == nil {
		log.Debug("Passing through message to counterparty")
		ret := Message{
			Name: MessagePassthruFailed,
		}
		c.out.Encode(ret)
	} else {
		c.Counterparty.out.Encode(msg)
	}
}

func (c *clientSession) RequestSendPhonon(msg Message) {

}

func (c *clientSession) RequestPhononAck(msg Message) {

}

func (c *clientSession) sendPhonon(msg Message) {
	// save this packet for later
	// delete after ack
}

func (c *clientSession) ack(msg Message) {
	// delete saved phonon when received
}
