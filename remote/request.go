package remote

type RequestType string

// haha ask me what the difference between a payload and parameters is
type Request struct{
	Name RequestType
	Payload string
}

var (
	RequestCardPair1 = "CardPair1"
	RequestCardPair2 = "CardPair2"
	RequestFinalizeCardPair = "FinalizeCardPair"
	RequestGetCertificate = "GetCertificate"
	RequestCardChallenge = "CardChallenge"
)

