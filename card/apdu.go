package card

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"

	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/globalplatform"
	"github.com/GridPlus/phonon-client/util"
	log "github.com/sirupsen/logrus"
)

const (
	maxAPDULength = 256

	InsIdentifyCard    = 0x14
	InsVerifyPIN       = 0x20
	InsChangePIN       = 0x21
	InsCreatePhonon    = 0x30
	InsSetDescriptor   = 0x31
	InsListPhonons     = 0x32
	InsGetPhononPubKey = 0x33
	InsDestroyPhonon   = 0x34
	InsSendPhonons     = 0x35
	InsRecvPhonons     = 0x36

	TagSelectAppInfo           = 0xA4
	TagCardUID                 = 0x8F
	TagCardSecureChannelPubKey = 0x80
	TagAppVersion              = 0x02
	TagPairingSlots            = 0x03
	TagAppCapability           = 0x8D

	TagPhononKeyCollection = 0x40
	TagKeyIndex            = 0x41
	TagPhononPubKey        = 0x80
	TagPhononPrivKey       = 0x81

	TagPhononFilter        = 0x60
	TagValueFilterLessThan = 0x84
	TagValueFilterMoreThan = 0x85

	TagPhononCollection = 0x52
	TagPhononDescriptor = 0x50
	TagPhononValue      = 0x83
	TagCurrencyType     = 0x81

	TagPhononKeyIndexList       = 0x42
	TagTransferPhononPacket     = 0x43
	TagPhononPrivateDescription = 0x44

	StatusSuccess         = 0x9000
	StatusPhononTableFull = 0x6A84
	StatusInvalidFile     = 0x6983
)

var (
	ErrCardUninitialized = errors.New("card uninitialized")
	ErrPhononTableFull   = errors.New("phonon table full")
	ErrUnknown           = errors.New("unknown error")
)

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
		[]byte{0x00},
	)
}

//TODO: implement with TLV encoding, fix for uint16 keyIndex return value
func ParseCreatePhononResponse(resp []byte) (keyIndex int, pubKey *ecdsa.PublicKey, err error) {
	log.Debug("create phonon response length: ", len(resp))
	collection, err := ParseTLVPacket(resp, TagPhononKeyCollection)
	if err != nil {
		return 0, nil, err
	}

	keyIndexBytes, err := collection.FindTag(TagKeyIndex)
	if err != nil {
		return 0, nil, err
	}

	pubKeyBytes, err := collection.FindTag(TagPhononPubKey)
	if err != nil {
		return 0, nil, err
	}

	keyIndex = int(binary.BigEndian.Uint16(keyIndexBytes))

	pubKey, err = util.ParseECDSAPubKey(pubKeyBytes)
	if err != nil {
		log.Error("could not parse pubkey from phonon response: ", err)
		return keyIndex, nil, err
	}

	return keyIndex, pubKey, nil
}

func NewCommandSetDescriptor(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsSetDescriptor,
		0x00,
		0x00,
		data,
	)
}

func NewCommandListPhonons(p1 byte, p2 byte, data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsListPhonons,
		p1,
		p2,
		data,
	)
}

func NewCommandGetPhononPubKey(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsGetPhononPubKey,
		0x00,
		0x00,
		data,
	)
}

func NewCommandDestroyPhonon(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsDestroyPhonon,
		0x00,
		0x00,
		data,
	)
}

func NewCommandSendPhonons(keyIndices []uint16, extendedRequest bool) *apdu.Command {
	var p1 byte
	if extendedRequest {
		p1 = 0x01
	} else {
		p1 = 0x00
	}

	p2 := byte(len(keyIndices))

	var keyIndexBytes []byte
	b := make([]byte, 2)
	for _, keyIndex := range keyIndices {
		binary.BigEndian.PutUint16(b, keyIndex)
		keyIndexBytes = append(keyIndexBytes, b...)
	}
	//TODO: possibly handle potential error
	data, _ := NewTLV(TagPhononKeyIndexList, keyIndexBytes)

	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsSendPhonons,
		p1,
		p2,
		data.Encode(),
	)
}

//Receives a TLV encoded Phonon Transfer Packet Payload in encrypted form
//and passes it on directly to a card
func NewCommandReceivePhonons(phononTransferPacket []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsRecvPhonons,
		0x00,
		0x00,
		phononTransferPacket,
	)
}
