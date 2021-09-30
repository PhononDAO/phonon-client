package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
)

type ConnectBridgeConnection struct {
	connection *websocket.Conn
}

func main(){
	bridge, err := ConnectToBridge("ws://localhost:8080", "nate", "otherguy")
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	bridge.ReadFromServer()
}

func ConnectToBridge(bridgeUrl string, yourName string, counterPartyName string) (ConnectBridgeConnection, error) {
	dialer := websocket.Dialer{}
	headers := http.Header{}
	conn,resp, err := dialer.Dial(fmt.Sprintf("%s/websocket",bridgeUrl),headers)
	if err != nil{
		//hopefully there's no error during handling the error
		resptext, readerr := ioutil.ReadAll(resp.Body)
		if readerr != nil{
			resptext = []byte("")
		}
		return ConnectBridgeConnection{}, fmt.Errorf("Unable to establish websocket connection to bridge: %s\nServer Response: %s", err.Error(),string(resptext))
	}
	return ConnectBridgeConnection{
		connection: conn,
	}, nil
}

type message string

func (c *ConnectBridgeConnection)ReadFromServer(){
	var mesg message
	c.connection.ReadJSON(&mesg)
	fmt.Println(mesg)
}
/*
func (c *ConnectBridgeConnection) GetCertificate() (cert.CardCertificate, error)
func (c *ConnectBridgeConnection) CardPair(initPairingData []byte) (cardPairData []byte, err error)
func (c *ConnectBridgeConnection) CardPair2(cardPairData []byte) (cardPairData2 []byte, err error)
func (c *ConnectBridgeConnection) FinalizeCardPair(cardPair2Data []byte) error
func (c *ConnectBridgeConnection) ReceivePhonons(phononTransfer []byte) error
func (c *ConnectBridgeConnection) RequestPhonons(phonons []model.Phonon) (phononTransfer []byte, err error)
func (c *ConnectBridgeConnection) GenerateInvoice() (invoiceData []byte, err error)
func (c *ConnectBridgeConnection) ReceiveInvoice(invoiceData []byte) error
*/
