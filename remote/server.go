package remote

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
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
	err := http.ListenAndServeTLS("localhost:"+port, certfile, keyfile, nil)
	if err != nil {
		fmt.Printf("Error with web server:, %s", err.Error())
	}
}

type clientSession struct {
	certificate    *cert.CardCertificate
	challengeNonce [32]byte
	underlyingConn *h2conn.Conn
	sender         *gob.Encoder
	receiver       *gob.Decoder
	validated      bool
	end            chan bool
	counterparty   *clientSession
	// the same name that goes in the lookup value of the clientSession map
	name string
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
		log.Debug("Unable to establish http2 duplex connection with ", r.RemoteAddr)
		//status teapot is obviously wrong here. need to research what causes this error and return a proper response
		http.Error(w, "Unable to establish duplex connection between server and client", http.StatusTeapot)
	}
	cmdEncoder := gob.NewEncoder(conn)
	cmdDecoder := gob.NewDecoder(conn)
	//generate session
	session := clientSession{
		sender:   cmdEncoder,
		receiver: cmdDecoder,
		end:      make(chan (bool)),
	}

	messageChan := make(chan (Message))

	go func(msgchan chan Message) {
		defer close(msgchan)
		for {
			message := Message{}
			err := session.receiver.Decode(&message)
			if err != nil {
				log.Info("Error receiving message from connected client")
				return
			}
			msgchan <- message
		}
	}(messageChan)

	newMessage := Message{
		Name: RequestCertificate,
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

func (c *clientSession) process(msg Message) {
	// if the client hasn't identified itself with the server, ignore what they are doing until they provide the certificate, and keep asking for it.
	fmt.Println("processing")
	if c.certificate == nil {
		// if they are providing the certificate, accept it, and then generate a challenge, add it to the challenge test, and continue executing
		if msg.Name == ResponseCertificate {
			fmt.Println("gettingCertificate")
			certParsed, err := cert.ParseRawCardCertificate(msg.Payload)
			if err != nil {
				log.Info("failed to parse certificate from client %s", err.Error())
				return
			}
			c.certificate = &certParsed
		} else {
			//ask for the certificate again
			c.sender.Encode(Message{
				Name: RequestCertificate,
			})
			return
		}
	}

	if !c.validated {
		if msg.Name == ResponseIdentify {
			fmt.Println("processing identify command")
			fmt.Println("getting IdentifyCard")
			buf := bytes.NewBuffer(msg.Payload)
			decoder := gob.NewDecoder(buf)
			var sig = &util.ECDSASignature{}
			err := decoder.Decode(sig)
			if err != nil {
				log.Error("Unable to parse IdentifyCardResponse", err.Error())
				return
			}
			key, err := util.ParseECDSAPubKey(c.certificate.PubKey)
			if err != nil {
				log.Error("Unable to parse pubkey from certificate", err.Error())
				return
			}
			fmt.Println(sig.R, sig.S)
			fmt.Println(c.challengeNonce[:])
			fmt.Println(key)
			if !ecdsa.Verify(key, c.challengeNonce[:], sig.R, sig.S) {
				log.Error("unable to verify card challenge")
				return
			}
			c.validated = true
			//todo: get the short name the same way it works in the repl
			clientSessions[string(c.certificate.Sig)] = c
			return
			//if challenge text wasn't set, set it, and send the challenge to the card
		} else {
			fmt.Println("generating card challenge")
			if c.challengeNonce == [32]byte{} {
				_, err := rand.Reader.Read(c.challengeNonce[:])
				if err != nil {
					fmt.Println("wtf")
					return
				}
			}
			c.RequestIdentify()
			return
		}
		// if the challenge text has been set, ignore what they want and send it again
	}
	switch msg.Name {
	case RequestConnectCard2Card:
		c.ConnectCard2Card(msg)
		//connect2card
		// check to see if a connected card has the ID
		// associate passthru with the opposing card
	case RequestDisconnectFromCard:
		c.DisconnectFromCard(msg)
		//disconnectFromCard
		// unassociate the cards
	case RequestEndSession:
		c.EndSession(msg)
		//endSession
		// end session
	case RequestNoOp:
		c.noop(msg)
		//noop
		// do nothing, but reset the last communication counter
	case ResponseIdentify, RequestCardPair1, RequestCardPair2, RequestFinalizeCardPair:
		c.passthrough(msg)
		//passthru msg
		// send to the opposing card
	case RequestSendPhonon:
		c.sendPhonon(msg)
		//passthru phonon packet
		// generate uuid
		// write packet to pending phonon table with card id and phonon uuid
		// send packet to opposing card
	case RequestPhononAck:
		c.ack(msg)
		//phonon ack
		// read uuid
		// delete from pending phonon table

	}
	fmt.Printf("%+v", msg)
}

func (c *clientSession) RequestIdentify() {
	c.sender.Encode(Message{
		Name:    RequestIdentify,
		Payload: c.challengeNonce[:],
	})
}

func (c *clientSession) ProvideCertificate() {
	msg := Message{
		Name:    ResponseCertificate,
		Payload: c.counterparty.certificate.Serialize(),
	}
	err := c.sender.Encode(msg)
	if err != nil {
		log.Error("Error encoding provideCertificate reply: ", err)
		return
	}
	return
}

func (c *clientSession) ConnectCard2Card(msg Message) {
	counterparty, ok := clientSessions[string(msg.Payload)]
	if !ok {
		c.sender.Encode(Message{
			Name:    MessageError,
			Payload: []byte("No connected card"),
		})
		log.Error("No connected session:", string(msg.Payload))
		return
	} else if counterparty.counterparty == nil && c.counterparty == nil {
		counterparty.counterparty = c
		c.counterparty = counterparty
		c.sender.Encode(Message{
			Name: MessageConnected,
			// Send back the name of the person you've connected to
			Payload: msg.Payload,
		})
		c.counterparty.sender.Encode(Message{
			Name:    MessageConnected,
			Payload: []byte(c.name),
		})
	} else {
		c.sender.Encode(Message{
			Name:    MessageError,
			Payload: []byte("Unable to connect. Connection already satisfied"),
		})
	}
}

func (c *clientSession) DisconnectFromCard(msg Message) {
	out := Message{
		Name: MessageDisconnected,
	}
	// encode can fail, so it needs to be checked. Not sure how to handle that
	c.counterparty.sender.Encode(out)
	c.sender.Encode(out)
	c.counterparty.counterparty = nil
	c.counterparty = nil
}

func (c *clientSession) EndSession(msg Message) {
	if c.counterparty != nil {
		c.DisconnectFromCard(msg)
	}
	c.underlyingConn.Close()
}

func (c *clientSession) noop(msg Message) {
	// don't do anything
	// this is eventually going to be for preventing connection timeouts, but may not be nessesary in the future
}

func (c *clientSession) passthrough(msg Message) {
	if c.counterparty == nil {
		ret := Message{
			Name: MessagePassthruFailed,
		}
		c.sender.Encode(ret)
		return
	}
	c.counterparty.sender.Encode(msg)
	// needs error handling on the encoding
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
