package card

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/GridPlus/keycard-go"
	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/globalplatform"
	"github.com/GridPlus/keycard-go/gridplus"
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
	InsGenerateInvoice  = 0x54
	InsReceiveInvoice   = 0x55
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

	TagPhononCollection      = 0x52
	TagPhononDescriptor      = 0x50
	TagPhononValue           = 0x83
	TagCurrencyType          = 0x82
	TagCurveType             = 0x87
	TagSchemaVersion         = 0x88
	TagExtendedSchemaVersion = 0x89

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
	TagAESKey          = 0x95

	TagInvoiceID = 0x96

	//ISO7816 Standard Responses
	SW_APPLET_SELECT_FAILED           = 0x6999
	SW_BYTES_REMAINING_00             = 0x6100
	SW_CLA_NOT_SUPPORTED              = 0x6E00
	SW_COMMAND_CHAINING_NOT_SUPPORTED = 0x6884
	SW_COMMAND_NOT_ALLOWED            = 0x6986
	SW_CONDITIONS_NOT_SATISFIED       = 0x6985
	SW_CORRECT_LENGTH_00              = 0x6C00
	SW_DATA_INVALID                   = 0x6984
	SW_FILE_FULL                      = 0x6A84
	SW_FILE_INVALID                   = 0x6983
	SW_FILE_NOT_FOUND                 = 0x6A82
	SW_FUNC_NOT_SUPPORTED             = 0x6A81
	SW_INCORRECT_P1P2                 = 0x6A86
	SW_INS_NOT_SUPPORTED              = 0x6D00
	SW_LAST_COMMAND_EXPECTED          = 0x6883
	SW_LOGICAL_CHANNEL_NOT_SUPPORTED  = 0x6881
	SW_NO_ERROR                       = 0x9000
	SW_RECORD_NOT_FOUND               = 0x6A83
	SW_SECURE_MESSAGING_NOT_SUPPORTED = 0x6882
	SW_SECURITY_STATUS_NOT_SATISFIED  = 0x6982
	SW_UNKNOWN                        = 0x6F00
	SW_WARNING_STATE_UNCHANGED        = 0x6200
	SW_WRONG_DATA                     = 0x6A80
	SW_WRONG_LENGTH                   = 0x6700
	SW_WRONG_P1P2                     = 0x6B00
)

type Command struct {
	ApduCmd      *apdu.Command
	PossibleErrs CardErrors
}

type CardErrors map[int]string

func (cmd *Command) HumanReadableErr(res *apdu.Response) error {
	var ret error
	errormsg, exists := cmd.PossibleErrs[int(res.Sw)]
	if exists {
		ret = fmt.Errorf(errormsg)
	}
	return ret
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
			SW_DATA_INVALID:      "Invalid Data Received",
			SW_INS_NOT_SUPPORTED: "Invalid INS",
		},
	}
}

func NewCommandVerifyPIN(pin string) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsVerifyPIN,
			0,
			0,
			[]byte(pin),
		),
		PossibleErrs: map[int]string{
			0x63c: "Pin Verification Failed",
		},
	}
}

func NewCommandChangePIN(pin string) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsChangePIN,
			0,
			0,
			[]byte(pin),
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_INCORRECT_P1P2:           "Invalid parameters",
		},
	}
}

func NewCommandCreatePhonon(curveType byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsCreatePhonon,
			curveType,
			0x00,
			[]byte{0x00},
		),
		PossibleErrs: map[int]string{
			SW_FILE_FULL:                "Phonon table full",
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
		},
	}
}

func NewCommandSetDescriptor(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsSetDescriptor,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_WRONG_LENGTH:             "Wrong data length",
			SW_FILE_INVALID:             "Invalid data",
			SW_FUNC_NOT_SUPPORTED:       "Phonon type not supported",
		},
	}
}

func NewCommandListPhonons(p1 byte, p2 byte, data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsListPhonons,
			p1,
			p2,
			data,
		),
		PossibleErrs: map[int]string{
			SW_WRONG_DATA:               "Request invalid",
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_INCORRECT_P1P2:           "Incorrect Parameters received",
		},
	}
}

func NewCommandGetPhononPubKey(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsGetPhononPubKey,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_WRONG_LENGTH:             "Data length incorrect",
			SW_WRONG_DATA:               "Phonon index invalid",
			SW_FILE_INVALID:             "Phonon Invalid",
			SW_FILE_NOT_FOUND:           "Phonon not initialized",
		},
	}
}

func NewCommandDestroyPhonon(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsDestroyPhonon,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_WRONG_LENGTH:             "Incoming length wrong",
			SW_WRONG_DATA:               "Invalid phonon index",
			SW_FILE_INVALID:             "Invalid phonon index",
		},
	}
}

func NewCommandSendPhonons(data []byte, p2Length byte, extendedRequest bool) *Command {
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
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_INCORRECT_P1P2:           "Incorrect parameters sent",
			SW_WRONG_DATA:               "Incorrect phonon index",
		},
	}
}

//Receives a TLV encoded Phonon Transfer Packet Payload in encrypted form
//and passes it on directly to a card
func NewCommandReceivePhonons(phononTransferPacket []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsRecvPhonons,
			0x00,
			0x00,
			phononTransferPacket,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Phonon recipt conditions not met",
			SW_FILE_FULL:                "Maximum number of phonons reached",
			SW_WRONG_DATA:               "Phonon received not valid",
		},
	}
}

func NewCommandSetReceiveList(data []byte) *Command {

	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsSetRecvList,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin not validated",
			SW_FILE_FULL:                "No phonon with index passed",
			SW_WRONG_DATA:               "Phonon to be received not valid",
		},
	}
}

func NewCommandTransactionAck(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaISO7816,
			InsTransactionAck,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED: "Pin Not Validated",
			SW_WRONG_DATA:               "Wrong data supplied"},
	}
}

func NewCommandInitCardPairing(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsInitCardPairing,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{},
	}
}

func NewCommandCardPair(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsCardPair,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{

			SW_CONDITIONS_NOT_SATISFIED: "Pairing already completed",
			SW_INCORRECT_P1P2:           "Unable to determine pairing step",
		},
	}
}

func NewCommandCardPair2(data []byte) *Command {
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

func NewCommandFinalizeCardPair(data []byte) *Command {
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

func NewCommandInstallCert(data []byte) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsLoadCert,
			0x00,
			0x00,
			data,
		),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED:       "Secure Channel is already Open",
			SW_INCORRECT_P1P2:                 "Invalid P1 Parameter(Wrong instruction step)",
			SW_WRONG_LENGTH:                   "Invalid Data Length",
			SW_WRONG_DATA:                     "Invalid data",
			SW_SECURE_MESSAGING_NOT_SUPPORTED: "Certificate Not Loaded",
			SW_SECURITY_STATUS_NOT_SATISFIED:  "Unable to generate secret or Challenge failed. Unable to verify cryptogram",
		},
	}
}

// put here to be next to the select applet command function
var phononAID = []byte{0xA0, 0x00, 0x00, 0x08, 0x20, 0x00, 0x03, 0x01}

func NewCommandSelectPhononApplet() *Command {
	return &Command{
		ApduCmd: globalplatform.NewCommandSelect(phononAID),
		// no errors known
		PossibleErrs: map[int]string{},
	}
}

func NewCommandPairStep1(salt []byte, pairingPubKey *ecdsa.PublicKey) *Command {
	return &Command{
		ApduCmd: gridplus.NewAPDUPairStep1(salt, pairingPubKey),
		PossibleErrs: map[int]string{
			SW_WRONG_DATA:                     "Data incorrect size",
			SW_SECURE_MESSAGING_NOT_SUPPORTED: "No certificate loaded",
			SW_SECURITY_STATUS_NOT_SATISFIED:  "Unable to compute ECDH secrets",
		},
	}

}

func NewCommandPairStep2(cryptogram [32]byte) *Command {
	return &Command{
		ApduCmd: gridplus.NewAPDUPairStep2(cryptogram[0:]),
		PossibleErrs: map[int]string{
			SW_WRONG_DATA:                    "Wrong secret length",
			SW_SECURITY_STATUS_NOT_SATISFIED: "Client cryptogram differs from expected",
		},
	}

}

func NewCommandUnpair(index uint8) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandUnpair(index),
		// No errors known
		PossibleErrs: map[int]string{},
	}

}

func NewCommandOpenSecureChannel(index uint8, publicKey []byte) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandOpenSecureChannel(index, publicKey),
		PossibleErrs: map[int]string{
			SW_INCORRECT_P1P2:                "Incorrect parameters",
			SW_SECURITY_STATUS_NOT_SATISFIED: "Unable to generate secret",
		},
	}

}

func NewCommandMutualAuthenticate(data []byte) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandMutuallyAuthenticate(data),
		PossibleErrs: map[int]string{
			SW_CONDITIONS_NOT_SATISFIED:      "Authentication key not initialized",
			SW_LOGICAL_CHANNEL_NOT_SUPPORTED: "Already Mutually Authenticated",
			SW_SECURITY_STATUS_NOT_SATISFIED: "Secret length invalid",
		},
	}

}

func NewCommandInit(data []byte) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandInit(data),
		// No errors known
		PossibleErrs: map[int]string{},
	}

}

func NewCommandGenerateInvoice() *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsGenerateInvoice,
			0x00,
			0x00,
			[]byte{0x00},
		),
		//TODO: Errors
		PossibleErrs: map[int]string{},
	}
}

func NewCommandReceiveInvoice() *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsReceiveInvoice,
			0x00,
			0x00,
			[]byte{0x00},
		),
		//TODO: Errors
		PossibleErrs: map[int]string{},
	}
}
