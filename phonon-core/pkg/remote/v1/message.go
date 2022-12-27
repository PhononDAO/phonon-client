package v1

// haha ask me what the difference between a payload and parameters is
type Message struct {
	Name    string
	Payload []byte
}

var (
	// Server to client messages
	MessageConnected            = "Connected"
	MessageDisconnected         = "Disconnected"
	MessageError                = "Error"
	MessagePassthruFailed       = "PassthruFailed"
	MessageIdentifiedWithServer = "IdentifiedWithServer"
	MessageConnectedToCard      = "connectedToCard"

	// Client to server commands
	RequestIdentify           = "Identify"
	ResponseIdentify          = "IdentifyResponse"
	RequestCertificate        = "RequestCert"
	ResponseCertificate       = "ResponseCert"
	RequestNoOp               = "NoOp"
	RequestConnectCard2Card   = "Connect2Card"
	RequestDisconnectFromCard = "DisconnectFromCard"
	RequestEndSession         = "EndSession"
	MessagePhononAck          = "AckPhonon"

	// Client to client commands
	RequestVerifyPaired      = "VerifyPairing"
	ResponseVerifyPaired     = "VerifyPairingRespnose"
	RequestCardPair1         = "CardPair1"
	ResponseCardPair1        = "CardPair1Response"
	RequestCardPair2         = "CardPair2"
	ResponseCardPair2        = "CardPair2Response"
	RequestFinalizeCardPair  = "FinalizeCardPair"
	ResponseFinalizeCardPair = "FinalizeCardPairResponse"
	// this one is weird because the server will cache this one
	RequestReceivePhonon = "requestReceivePhonon"
)
