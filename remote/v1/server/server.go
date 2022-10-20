package server

import (
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GridPlus/phonon-client/cert"
	v1 "github.com/GridPlus/phonon-client/remote/v1"
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
	underlyingConn *h2conn.Conn
	out            *gob.Encoder
	in             *gob.Decoder
	Counterparty   *clientSession
	certificate    cert.CardCertificate
	validated      bool
	// the same name that goes in the lookup value of the clientSession map
}

var clientSessions map[string]*clientSession

func index(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("hello there"))
}

func listConnected(w http.ResponseWriter, _ *http.Request) {
	sess := make([]string, 0)
	for name, s := range clientSessions {
		sess = append(sess, fmt.Sprintf("session %s: %+v", name, s))
		sess = append(sess, fmt.Sprintf("counterparty: %+v", s.Counterparty))
	}
	json.NewEncoder(w).Encode(sess)
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

	valid, key, err := session.ValidateClient()
	if err != nil {
		err = session.out.Encode(err.Error())
		if err != nil {
			log.Error("failed sending cert validation failure response: ", err)
			return
		}
	}
	if !valid {
		//TODO: use a real error, possibly from cert package
		msg := v1.Message{
			Name:    v1.MessageDisconnected,
			Payload: []byte("Certificate invalid"),
		}
		err = session.out.Encode(msg)
		if err != nil {
			log.Error("failed sending invalid cert response: ", err)
			return
		}
		log.Error("certificate invalid")
	}

	name := util.CardIDFromPubKey(key)
	session.Name = strings.ToLower(name)
	clientSessions[name] = &session
	session.out.Encode(v1.Message{
		Name:    v1.MessageIdentifiedWithServer,
		Payload: []byte(name),
	})

	log.Info("validated client connection: ", session.Name)
	defer session.endSession(v1.Message{})
	//Client is now validated, move on
	for r.Context().Err() == nil {
		var msg v1.Message
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

func (c *clientSession) process(msg v1.Message) error {
	switch msg.Name {
	case v1.RequestConnectCard2Card:
		c.ConnectCard2Card(msg)
	case v1.RequestDisconnectFromCard:
		c.disconnectFromCard(msg)
	case v1.RequestEndSession:
		c.endSession(msg)
	case v1.RequestNoOp:
		c.noop(msg)
	case v1.RequestCertificate:
		c.provideCertificate()
	default:
		fmt.Printf("passing through %+v\n", msg)
		c.passthrough(msg)
	}
	//TODO: provide actual errors, or ensure all the cases handle errors themselves
	return nil
}

func (c *clientSession) ValidateClient() (bool, *ecdsa.PublicKey, error) {
	log.Info("validating client connection")
	//Read client certificate
	var in v1.Message
	err := c.in.Decode(&in)
	if err != nil {
		log.Error("unable to decode raw client certificate bytes: ", err)
		return false, nil, err
	}
	log.Info("past first Decode:")
	c.certificate, err = cert.ParseRawCardCertificate(in.Payload)
	if err != nil {
		log.Infof("failed to parse certificate from client %s\n", err.Error())
		return false, nil, err
	}
	log.Info("parsed cert: ", c.certificate)
	//Validate certificate is signed by valid origin

	//Send Identify Card Challenge
	challengeNonce, err := c.RequestIdentify()
	if err != nil {
		log.Error("failed to send IDENTIFY_CARD request: ", err)
		return false, nil, nil
	}

	sig, err := c.ReceiveIdentifyResponse()
	if err != nil {
		log.Error("failed to receive IDENTIFY_CARD response: ", err)
		return false, nil, err
	}
	log.Infof("received sig from identifyResponse: %+v", sig)
	key, err := util.ParseECCPubKey(c.certificate.PubKey)
	if err != nil {
		log.Error("Unable to parse pubkey from certificate", err.Error())
		return false, nil, err
	}
	if !ecdsa.Verify(key, challengeNonce, sig.R, sig.S) {
		log.Error("unable to verify card challenge")
		return false, nil, err
	}
	//Cert has been validated, register clientSession with server and grab card name
	c.validated = true
	//Return to main loop to process further client requests
	return true, key, nil
}
