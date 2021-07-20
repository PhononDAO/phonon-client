package card

import (
	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/globalplatform"
)

const (
	maxAPDULength = 256

	InsIdentifyCard     = 0x14
	InsVerifyPIN        = 0x20
	InsChangePIN        = 0x21
	InsCreatePhonon     = 0x30
	InsSetDescriptor    = 0x31
	InsListPhonons      = 0x32
	InsGetPhononPubKey  = 0x33
	InsDestroyPhonon    = 0x34
	InsSendPhonons      = 0x35
	InsRecvPhonons      = 0x36
	InsSetRecvList      = 0x37
	InsTransactionAck   = 0x38
	InsInitCardPairing  = 0x50
	InsCardPair         = 0x51
	InsCardPair2        = 0x52
	InsFinalizeCardPair = 0x53
	InsInstallKey       = 0x15

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

	TagPhononPubKeyList = 0x7F

	TagCardCertificate = 0x90
	TagECCPublicKey    = 0x80 //TODO: resolve redundancy around 0x80 tag
	TagSalt            = 0x91
	TagAesIV           = 0x92
	TagECDSASig        = 0x93
	TagPairingIndex    = 0x94
)

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

func NewCommandSendPhonons(data []byte, p2Length byte, extendedRequest bool) *apdu.Command {
	var p1 byte
	if extendedRequest {
		p1 = 0x01
	} else {
		p1 = 0x00
	}

	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsSendPhonons,
		p1,
		p2Length,
		data,
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

func NewCommandSetReceiveList(data []byte) *apdu.Command {

	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsSetRecvList,
		0x00,
		0x00,
		data,
	)
}

func NewCommandTransactionAck(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaISO7816,
		InsTransactionAck,
		0x00,
		0x00,
		data,
	)
}

func NewCommandInitCardPairing() *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsInitCardPairing,
		0x00,
		0x00,
		nil,
	)
}

func NewCommandCardPair(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsCardPair,
		0x00,
		0x00,
		data,
	)
}

func NewCommandCardPair2(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsCardPair2,
		0x00,
		0x00,
		data,
	)
}

func NewCommandFinalizeCardPair(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsFinalizeCardPair,
		0x00,
		0x00,
		data,
	)
}

func NewCommandInstallCert(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsInstallKey,
		0x00,
		0x00,
		data,
	)
}

