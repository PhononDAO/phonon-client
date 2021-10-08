package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/posener/h2conn"
	log "github.com/sirupsen/logrus"
)

func StartServer(port string, certfile string, keyfile string) {
	http.HandleFunc("/phonon", handle)
	http.HandleFunc("/connectec", listConnected)
	http.HandleFunc("/", index)
	err := http.ListenAndServeTLS("localhost:"+port, certfile, keyfile, nil)
	if err != nil {
		fmt.Printf("Error with web server:, %s", err.Error())
	}
}

type clientSession struct {
	certificate   *cert.CardCertificate
	challengeText string
	sender        *json.Encoder
	receiver      *json.Decoder
	validated     bool
	end           chan bool
	counterparty  *clientSession
}

var clientSessions map[string]clientSession

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello there"))
}

func listConnected(w http.ResponseWriter, r *http.Request) {
	connected := fmt.Sprintf("%+v", clientSessions)
	w.Write([]byte(connected))
}

func handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h2conn.Accept(w, r)
	if err != nil {
		log.Debug("Unable to establish http2 duplex connection with ", r.RemoteAddr)
		//status teapot is obviously wrong here. need to research what causes this error and return a proper response
		http.Error(w, "Unable to establish duplex connection between server and client", http.StatusTeapot)
	}
	cmdEncoder := json.NewEncoder(conn)
	cmdDecoder := json.NewDecoder(conn)
	//generate session
	session := clientSession{
		sender:   cmdEncoder,
		receiver: cmdDecoder,
		end:      make(chan (bool)),
	}

	messageChan := make(chan (Request))

	go func(msgchan chan Request) {
		defer close(msgchan)
		for {
			message := Request{}
			err := session.receiver.Decode(&message)
			if err != nil {
				log.Info("Error receiving message from connected client")
				return
			}
			msgchan <- message
		}
	}(messageChan)

	newMessage := Request{
		Name:    RequestCardChallenge,
		Payload: "asdbfasjdklf",
	}

	cmdEncoder.Encode(newMessage)

	for message := range messageChan {
		session.process(message)
	}
	conn.Close()

	//ask connected card for a certificate and send challenge
	//when certificate is verified, add to list of connected cards
	//process messages
}

//this is unironically some of the worst code I've ever written
func (c *clientSession) process(req Request) {
	// if the client hasn't identified itself with the server, ignore what they are doing until they provide the certificate, and keep asking for it.
	if c.certificate == nil {
		// if they are providing the certificate, accept it, and then generate a challenge, add it to the challenge test, and continue executing
		if req.Name == ResponseProvideCertificate {
			rawCert, err := base64.StdEncoding.DecodeString(req.Payload)
			if err != nil {
				//todo add client information
				log.Info("failed to parse message from client connected")
				return
			}
			certParsed, err := cert.ParseRawCardCertificate(rawCert)
			if err != nil {
				log.Info("failed to parse certificate from client %s", err.Error())
				return
			}
			c.certificate = &certParsed
		} else {
			//ask for the certificate again
			c.sender.Encode(Request{
				Name: ResponseProvideCertificate,
			})
			return
		}
	}
	if !c.validated {
		if req.Name == ResponseCardChallenge {
			// parse challenge response
			// check if response is the same as the challenge
			// add card to list of active cards
			// set validated == true
			return
			//if challenge text wasn't set, set it, and send the challenge to the card
		} else {
			if c.challengeText == "" {
				//generate challengeText
			}
			// generate challenge
			// send it off
			return
		}
		// if the challenge text has been set, ignore what they want and send it again
	}

	switch req.Name {
	case RequestConnectCard2Card:
		c.ConnectCard2Card(req)
		//connect2card
		// check to see if a connected card has the ID
		// associate passthru with the opposing card
	case RequestDisconnectFromCard:
		c.DisconnectFromCard(req)
		//disconnectFromCard
		// unassociate the cards
	case RequestEndSession:
		c.EndSession(req)
		//endSession
		// end session
	case RequestNoOp:
		c.noop(req)
		//noop
		// do nothing, but reset the last communication counter
	case RequestCardPair1, RequestCardPair2, RequestFinalizeCardPair:
		c.passthrough(req)
		//passthru msg
		// send to the opposing card
	case RequestSendPhonon:
		c.sendPhonon(req)
		//passthru phonon packet
		// generate uuid
		// write packet to pending phonon table with card id and phonon uuid
		// send packet to opposing card
	case RequestPhononAck:
		c.ack(req)
		//phonon ack
		// read uuid
		// delete from pending phonon table

	}
	fmt.Printf("%+v", req)
}

func (c *clientSession) ConnectCard2Card(Request) {

}
func (c *clientSession) DisconnectFromCard(Request) {

}

func (c *clientSession) EndSession(Request) {

}

func (c *clientSession) noop(Request) {

}

func (c *clientSession) passthrough(Request) {

}

func (c *clientSession) RequestSendPhonon(Request) {

}

func (c *clientSession) RequestPhononAck(Request) {

}

func (c *clientSession) sendPhonon(Request) {

}

func (c *clientSession) ack(Request) {

}
