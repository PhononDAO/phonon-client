package card

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

var ErrCardUninitialized = errors.New("card uninitialized")

func ParseSelectResponse(resp []byte) (instanceUID []byte, cardPubKey []byte, err error) {
	if len(resp) == 0 {
		return nil, nil, errors.New("received nil response")
	}
	switch resp[0] {
	//Initialized
	case 0xA4:
		log.Debug("card wallet initialized")
		//If length of length is set this is a long format TLV response
		if len(resp) < 88 {
			log.Error("response should have been at least length 86 bytes, was length: ", len(resp))
			return nil, nil, errors.New("invalid response length")
		}
		if resp[3] == 0x81 {
			instanceUID = resp[6:22]
			cardPubKey = resp[24:89]
		} else {
			instanceUID = resp[5:21]
			cardPubKey = resp[23:88]
		}
	case 0x80:
		log.Debug("card wallet uninitialized")
		length := int(resp[1])
		cardPubKey = resp[2 : 2+length]
		return nil, cardPubKey, ErrCardUninitialized
	}

	return instanceUID, cardPubKey, nil
}
