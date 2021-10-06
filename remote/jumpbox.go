package remote

import (
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
	certificate cert.CardCertificate
	sender      *json.Encoder
	receiver    *json.Decoder
	end         chan bool
}

var clientSessions map[string]clientSession

func index (w http.ResponseWriter, r *http.Request) {
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

	for message := range messageChan {
		session.process(message)
	}
	conn.Close()

	//ask connected card for a certificate and send challenge
	//when certificate is verified, add to list of connected cards
	//process messages
	//connect2card
	// check to see if a connected card has the ID
	// associate passthru with the opposing card
	//disconnectFromCard
	// unassociate the cards
	//endSession
	// end session
	//noop
	// do nothing, but reset the last communication counter
	//passthru msg
	// send to the opposing card
	//passthru phonon packet
	// generate uuid
	// write packet to pending phonon table with card id and phonon uuid
	// send packet to opposing card
	//phonon ack
	// read uuid
	// delete from pending phonon table
}

func (c *clientSession) process(message interface{}) {
	fmt.Println(message)
}
