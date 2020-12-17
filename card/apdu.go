package card

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/keycard-go/apdu"
	log "github.com/sirupsen/logrus"
)

//Manually parse possible TLV responses
func parseSelectResponse(resp []byte) (instanceUID []byte, cardPubKey []byte, err error) {
	if len(resp) == 0 {
		return nil, nil, errors.New("received nil response")
	}
	switch resp[0] {
	//Initialized
	case 0xA4:
		log.Info("card wallet initialized")
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
		log.Error("card wallet uninitialized")
		return nil, nil, errors.New("card wallet uninitialized")
	}

	return instanceUID, cardPubKey, nil
}

// func (c Safecard) SendPairStep1(clientSalt []byte, pubKey *ecdsa.PublicKey) (SafecardRAPDUStep1, error) {
// 	cmd := NewAPDUPairStep1(clientSalt, pubKey.X.Bytes(), pubKey.Y.Bytes())

// 	cmd.SetLe(0)
// 	rawCmd, err := cmd.Serialize()
// 	if err != nil {
// 		log.Error("could not serialize pair step 1 apdu. err: ", err)
// 		return SafecardRAPDUStep1{}, err
// 	}
// 	resp, err := c.card.Transmit(rawCmd)
// 	if err != nil {
// 		log.Error("could not send pair step 1 command. err: ", err)
// 		return SafecardRAPDUStep1{}, err
// 	}
// 	log.Infof("resp pair step 1:\n%s", hex.Dump(resp))
// 	pairStep1Resp, err := parsePairStep1Response(resp)
// 	if err != nil {
// 		log.Error("error parsing pairStep1 response. err: ", err)
// 		return SafecardRAPDUStep1{}, err
// 	}
// 	return pairStep1Resp, nil
// }

// func (c Safecard) sendPairStep2(cryptogram []byte) (SafecardRAPDUStep2, error) {
// 	cmd := newAPDUPairStep2(cryptogram)

// 	resp, err := c.c.Send(cmd)
// 	if err != nil {
// 		log.Error("unable to send pairStep2 command. err: ", err)
// 	}
// 	pairStep2Resp, err := parsePairStep2Response(resp.Data)
// 	if err != nil {
// 		log.Error("unable to parse pairStep2Response. err: ", err)
// 	}
// 	return pairStep2Resp, nil
// }

func newAPDUPairStep2(cryptogram []byte) *apdu.Command {
	return apdu.NewCommand(
		SAFECARD_APDU_CLA_ENCRYPTED_PROPRIETARY,
		SAFECARD_APDU_INS_PAIR,
		PAIR_STEP2,
		0x00,
		cryptogram,
	)
}

// func parsePairStep2Response(resp []byte) (SafecardRAPDUStep2, error) {
// 	log.Infof("raw pairStep2 resp: % X", resp)
// 	correctLength := 33
// 	if len(resp) != correctLength {
// 		log.Errorf("resp was length(%v). should have been length %v", len(resp), correctLength)
// 		return SafecardRAPDUStep2{}, errors.New("pairstep2 response was invalid length")
// 	}
// 	return SafecardRAPDUStep2{
// 		pairingIdx: int(resp[0]),
// 		salt:       resp[1:33],
// 	}, nil
// }

func NewAPDUOpenSecureChannel(pubKey ecdsa.PublicKey) *apdu.Command {
	payload := SerializePubKey(pubKey)
	return apdu.NewCommand(
		SAFECARD_APDU_CLA_ENCRYPTED_PROPRIETARY,
		SAFECARD_APDU_INS_OPEN_SECURE_CHANNEL,
		0x00,
		0x00,
		payload,
	)
}

func parseOpenSecureChannelResponse(resp []byte) (SafecardRAPDUOpenSecureChannel, error) {
	correctLength := 48
	if len(resp) != correctLength {
		log.Error("open secure channel response incorrect length. Should have been %v, was %v", correctLength, len(resp))
		return SafecardRAPDUOpenSecureChannel{}, errors.New("response invalid length")
	}
	return SafecardRAPDUOpenSecureChannel{
		salt:          resp[0:32],
		aesInitVector: resp[32:48],
	}, nil
}
