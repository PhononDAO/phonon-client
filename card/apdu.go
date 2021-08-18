package card

import (
	"fmt"

	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/globalplatform"
)

const (
	// general things
	maxAPDULength = 256

	// instructions
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
	InsLoadCert         = 0x15

	// tags
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


// Not exactly sure where this should go
type Command struct {
	ApduCmd      *apdu.Command
	PossibleErrs CardErrors
}

type CardErrors map[int]string


func(cmd *Command)HumanReadableErr(res *apdu.Response)error{
	var ret error
	errormsg, exists := cmd.PossibleErrs[int(res.Sw)]; if exists{
		ret = fmt.Errorf(errormsg)
	}
	return  ret
}

//NewCommandIdentifyCard takes a 32 byte nonce value and sends it along with the IDENTIFY_CARD APDU
//As a response it receives the card's public key and and a signature
//on the salt to prove posession of the private key
func NewCommandIdentifyCard(nonce []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsIdentifyCard,
			0,
			0,
			nonce,
		),
		PossibleErrs: map[int]string{
			0x6984: "Returned data is anot SHA256",
		},
	}
}

func NewCommandVerifyPIN(pin string)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsVerifyPIN,
			0,
			0,
			[]byte(pin),
		),
		PossibleErrs: map[int]string{
			//this one breaks the whole separation thing
			//Not sure how to reconcile this
			0x63c: "Pin Verification Failed",
		},
	}
}

func NewCommandChangePIN(pin string)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsChangePIN,
			0,
			0,
			[]byte(pin),
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandCreatePhonon()*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsCreatePhonon,
			0x00,
			0x00,
			[]byte{0x00},
		),
		PossibleErrs: map[int]string{
			0x6A84: "Phonon table full",
		},
	}
}

func NewCommandSetDescriptor(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsSetDescriptor,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			// find Key not found response code
		},
	}
}

func NewCommandListPhonons(p1 byte, p2 byte, data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsListPhonons,
			p1,
			p2,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandGetPhononPubKey(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsGetPhononPubKey,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandDestroyPhonon(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsDestroyPhonon,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandSendPhonons(data []byte, p2Length byte, extendedRequest bool)*Command {
	var p1 byte
	if extendedRequest {
		p1 = 0x01
	} else {
		p1 = 0x00
	}

	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsSendPhonons,
			p1,
			p2Length,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

//Receives a TLV encoded Phonon Transfer Packet Payload in encrypted form
//and passes it on directly to a card
func NewCommandReceivePhonons(phononTransferPacket []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsRecvPhonons,
			0x00,
			0x00,
			phononTransferPacket,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandSetReceiveList(data []byte)*Command {

	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsSetRecvList,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandTransactionAck(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsTransactionAck,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandInitCardPairing()*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsInitCardPairing,
			0x00,
			0x00,
			nil,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandCardPair(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsCardPair,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandCardPair2(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsCardPair2,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandFinalizeCardPair(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsFinalizeCardPair,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandInstallCert(data []byte)*Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsLoadCert,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			0x6985: "Secure Channel is already Open",
			0x6A86: "Invalid P1 Parameter(Wrong instruction step)",
			0x6A80: "Invalid Data Length",
			0x6882: "Certificate Not Loaded",
			0x6982: "Unable to generate secret or Challenge failed. Unable to verify cryptogram",
		},
	}
}
