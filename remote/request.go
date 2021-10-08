package remote

// haha ask me what the difference between a payload and parameters is
type Request struct {
	Name    string
	Payload string
}

var (
	// Client to terminal commands
	RequestProvideCertifcate   = "ProvideCert"
	ResponseProvideCertificate = "ProvideCertResponse"
	RequestCardChallenge       = "CardChallenge"
	ResponseCardChallenge      = "ChallengeResponse"
	RequestNoOp                = "NoOp"
	RequestConnectCard2Card    = "Connect2Card"
	RequestDisconnectFromCard  = "DisconnectFromCard"
	RequestEndSession          = "EndSession"
	RequestPhononAck           = "AckPhonon"

	// Client to client commands
	RequestCardPair1          = "CardPair1"
	ResponseCardPair1Response = "CardPair1Response"
	RequestCardPair2          = "CardPair2"
	ResponseRequestCardPair2  = "CardPair2Response"
	RequestFinalizeCardPair   = "FinalizeCardPair"
	ResponseCardPair1         = "FinalizeCardPairResponse"
	// this one is weird because the server will cache this one
	RequestSendPhonon = "SendPhonon"
)
