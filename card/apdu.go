package card

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/status-im/keycard-go/apdu"
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

func (c Safecard) SendPairStep1(clientSalt []byte, pubKey *ecdsa.PublicKey) (SafecardRAPDUStep1, error) {
	cmd := NewAPDUPairStep1(clientSalt, pubKey.X.Bytes(), pubKey.Y.Bytes())

	cmd.SetLe(0)
	rawCmd, err := cmd.Serialize()
	if err != nil {
		log.Error("could not serialize pair step 1 apdu. err: ", err)
		return SafecardRAPDUStep1{}, err
	}
	resp, err := c.card.Transmit(rawCmd)
	if err != nil {
		log.Error("could not send pair step 1 command. err: ", err)
		return SafecardRAPDUStep1{}, err
	}
	log.Infof("resp pair step 1:\n%s", hex.Dump(resp))
	pairStep1Resp, err := parsePairStep1Response(resp)
	if err != nil {
		log.Error("error parsing pairStep1 response. err: ", err)
		return SafecardRAPDUStep1{}, err
	}
	return pairStep1Resp, nil
}

func NewAPDUPairStep1(clientSalt []byte, pubKeyX []byte, pubKeyY []byte) *apdu.Command {
	log.Info("length of clientSalt: ", len(clientSalt))
	log.Infof("clientSalt: % X", clientSalt)
	//secp256k1 curve pub key.
	//Need to add uncompressed format byte in front of pub key
	var ECC_POINT_FORMAT_UNCOMPRESSED byte = 0x04

	pubKey := []byte{ECC_POINT_FORMAT_UNCOMPRESSED}
	pubKey = append(pubKey, pubKeyX...)
	pubKey = append(pubKey, pubKeyY...)

	payload := append(clientSalt, TLV_TYPE_CUSTOM, byte(len(pubKey)))
	payload = append(payload, pubKey...)
	log.Info("pubKey length: ", len(pubKey))
	log.Infof("pubKey: %X", pubKey)

	log.Info("payload length: ", len(payload))
	return apdu.NewCommand(
		SAFECARD_APDU_CLA_ENCRYPTED_PROPRIETARY,
		SAFECARD_APDU_INS_PAIR,
		PAIR_STEP1,
		0x00,
		payload,
	)
}

func parsePairStep1Response(resp []byte) (apduResp SafecardRAPDUStep1, err error) {
	// minSigLength := 72
	// maxSigLength := 74
	// certLength := 147 //8 TLV Header + 65 PubKey + 74 Sig
	// saltSize := 32

	// maxPairStep1RespLen := saltSize + certLength + maxSigLength
	// minPairStep1RespLen := saltSize + certLength + minSigLength
	// log.Debug("pair step 1 resp length: ", len(resp))
	// if len(resp) < minPairStep1RespLen || len(resp) > maxPairStep1RespLen {
	// 	log.Errorf("response was %v bytes. should have been %v bytes", len(resp), correctPairStep1RespLen)
	// 	return SafecardRAPDUStep1{}, errors.New("invalid pair step 1 response length")
	// }

	apduResp.safecardSalt = resp[0:32]
	certLength := int(resp[33])
	log.Info("cert length calculated at: ", certLength)
	//TODO: Will this break if the cert length changes?
	//The pubKey includes the TLV header for cardPubKey?
	//But the TLV tags are excluded from permissions...
	apduResp.safecardCert = SafecardCert{
		permissions: resp[34:38],                   //skip 2 byte TLV header, include 2 byte TLV field description
		pubKey:      resp[38 : 38+2+65],            //2 byte TLV, 65 byte pubkey
		sig:         resp[38+65+2 : 34+certLength], //sig can be 72 to 74 bytes
	}

	log.Infof("end of resp len(%v): % X", resp[34+certLength:])
	// sigLength := int(resp[34+certLength+1])
	apduResp.safecardSig = resp[34+certLength:]

	log.Infof("card salt length(%v):\n% X", len(apduResp.safecardSalt), apduResp.safecardSalt)
	log.Infof("card cert permissions length(%v):\n% X", len(apduResp.safecardCert.permissions), apduResp.safecardCert.permissions)
	log.Infof("card cert pubKey length(%v):\n% X", len(apduResp.safecardCert.pubKey), apduResp.safecardCert.pubKey)
	log.Infof("card cert sig length(%v):\n% X", len(apduResp.safecardCert.sig), apduResp.safecardCert.sig)

	log.Infof("card sig length(%v): % X", len(apduResp.safecardSig), apduResp.safecardSig)
	return apduResp, nil
}

func (c Safecard) sendPairStep2(cryptogram []byte) (SafecardRAPDUStep2, error) {
	cmd := newAPDUPairStep2(cryptogram)

	resp, err := c.c.Send(cmd)
	if err != nil {
		log.Error("unable to send pairStep2 command. err: ", err)
	}
	pairStep2Resp, err := parsePairStep2Response(resp.Data)
	if err != nil {
		log.Error("unable to parse pairStep2Response. err: ", err)
	}
	return pairStep2Resp, nil
}

func newAPDUPairStep2(cryptogram []byte) *apdu.Command {
	return apdu.NewCommand(
		SAFECARD_APDU_CLA_ENCRYPTED_PROPRIETARY,
		SAFECARD_APDU_INS_PAIR,
		PAIR_STEP2,
		0x00,
		cryptogram,
	)
}

func parsePairStep2Response(resp []byte) (SafecardRAPDUStep2, error) {
	log.Infof("raw pairStep2 resp: % X", resp)
	correctLength := 33
	if len(resp) != correctLength {
		log.Errorf("resp was length(%v). should have been length %v", len(resp), correctLength)
		return SafecardRAPDUStep2{}, errors.New("pairstep2 response was invalid length")
	}
	return SafecardRAPDUStep2{
		pairingIdx: int(resp[0]),
		salt:       resp[1:33],
	}, nil
}

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
