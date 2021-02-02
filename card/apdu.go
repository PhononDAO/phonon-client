package card

import (
	"errors"

	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/globalplatform"
	log "github.com/sirupsen/logrus"
)

const (
	InsIdentifyCard = 0x14
	InsVerifyPIN    = 0x20
	InsChangePIN    = 0x21
	InsCreatePhonon = 0x30
)

var ErrCardUninitialized = errors.New("card uninitialized")

func ParseSelectResponse(resp []byte) (instanceUID []byte, cardPubKey []byte, err error) {
	if len(resp) == 0 {
		return nil, nil, errors.New("received nil response")
	}
	log.Debug("length of select response data: ", len(resp))
	switch resp[0] {
	//Initialized
	case 0xA4:
		log.Debug("pin initialized")
		//If length of length is set this is a long format TLV response
		if len(resp) < 88 {
			log.Error("response should have been at least length 86 bytes, was length: ", len(resp))
			return nil, nil, errors.New("invalid response length")
		}
		instanceUID = resp[4:20]
		cardPubKey = resp[22:87]
		//Think this response pattern only existed when a safecard wallet was initialized
		// if resp[3] == 0x81 {
		// 	instanceUID = resp[6:22]
		// 	cardPubKey = resp[24:89]
		// } else {
		//This would be the existing parsing above
	case 0x80:
		log.Debug("pin uninitialized")
		length := int(resp[1])
		cardPubKey = resp[2 : 2+length]
		return nil, cardPubKey, ErrCardUninitialized
	}

	return instanceUID, cardPubKey, nil
}

//NewCommandIdentifyCard takes a 32 byte nonce value and sends it along with the IDENTIFY_CARD APDU
//As a response it receives the card's public key and and a signature
//on the salt to prove posession of the private key
func NewCommandIdentifyCard(nonce []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsIdentifyCard,
		0,
		0,
		nonce,
	)
}

func ParseIdentifyCardResponse(resp []byte) (cardPubKey []byte, sig []byte, err error) {
	correctLength := 67
	if len(resp) < 67 {
		log.Errorf("identify card response invalid length %v should be %v ", len(resp), correctLength)
		return nil, nil, err
	}
	cardPubKey = resp[2:67]
	sig = resp[67:]

	return cardPubKey, sig, nil
}

func NewCommandVerifyPIN(pin string) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsVerifyPIN,
		0,
		0,
		[]byte(pin),
	)
}

func NewCommandChangePIN(pin string) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsChangePIN,
		0,
		0,
		[]byte(pin),
	)
}

func NewCommandCreatePhonon() *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsCreatePhonon,
		0x00,
		0x00,
		nil,
	)
}
