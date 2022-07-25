package card

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/keycard-go"
	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/globalplatform"
	"github.com/GridPlus/keycard-go/gridplus"
)

const (
	// general things
	maxAPDULength = 256

	// instructions
	InsIdentifyCard       = 0x14
	InsLoadCert           = 0x15
	InsVerifyPIN          = 0x20
	InsChangePIN          = 0x21
	InsCreatePhonon       = 0x30
	InsSetDescriptor      = 0x31
	InsListPhonons        = 0x32
	InsGetPhononPubKey    = 0x33
	InsDestroyPhonon      = 0x34
	InsSendPhonons        = 0x35
	InsRecvPhonons        = 0x36
	InsSetRecvList        = 0x37
	InsTransactionAck     = 0x38
	InsInitCardPairing    = 0x50
	InsCardPair           = 0x51
	InsCardPair2          = 0x52
	InsFinalizeCardPair   = 0x53
	InsGenerateInvoice    = 0x54
	InsGetFriendlyName    = 0x56
	InsSetFriendlyName    = 0x57
	InsReceiveInvoice     = 0x55
	InsGetAvailableMemory = 0x99
	InsMineNativePhonon   = 0x41

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
	TagPhononDenomBase       = 0x83
	TagPhononDenomExp        = 0x86
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

	//extended tags
	TagChainID = 0x20

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
	SW_MINING_FAILED                  = 0x9001
	SW_PIN_VERIFY_FAILED              = 0x63c
)

var (
	ErrMiningFailed       = errors.New("native phonon mine attempt failed")
	ErrInvalidPhononIndex = errors.New("invalid phonon index")
	ErrDefault            = errors.New("unspecified error for command")
)

type Command struct {
	ApduCmd      *apdu.Command
	PossibleErrs CmdErrTable
}

type CmdErrTable map[int]error

func (cmd *Command) HumanReadableErr(res *apdu.Response) error {
	err, exists := cmd.PossibleErrs[int(res.Sw)]
	if exists {
		return err
		//Return unspecified error if code is not listed in command and is != 0x9000
	} else if res.Sw != SW_NO_ERROR {
		return ErrDefault
	}
	return nil
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
		PossibleErrs: CmdErrTable{
			SW_DATA_INVALID: errors.New("received Challenge is not correct length"),
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
		PossibleErrs: CmdErrTable{
			SW_PIN_VERIFY_FAILED: errors.New("pin verification failed"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_INCORRECT_P1P2:           errors.New("parameter neither change user pin or change pairing secret"),
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
		PossibleErrs: CmdErrTable{
			SW_FILE_FULL:                ErrPhononTableFull,
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_LENGTH:             errors.New("wrong data length"),
			SW_FILE_INVALID:             ErrInvalidPhononIndex,
			SW_FILE_INVALID + 1:         errors.New("phonon does not exist"),
			SW_FILE_INVALID + 3:         errors.New("phonon does not exist"),
			SW_FILE_INVALID + 4:         errors.New("unable to decode Currency TLV"),
			SW_FILE_INVALID + 5:         errors.New("unable to set currency type to 0x00"),
			SW_FILE_INVALID + 6:         errors.New("unable to decode Phonon Value TLV"),

			SW_FUNC_NOT_SUPPORTED: errors.New("phonon type not supported"),
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
		PossibleErrs: CmdErrTable{
			SW_WRONG_DATA:               errors.New("no remaining phonons to list"),
			SW_WRONG_DATA + 1:           errors.New("unable to decode phonon filter TLV"),
			SW_WRONG_DATA + 2:           errors.New("unable to decode phonon currency TLV"),
			SW_WRONG_DATA + 3:           errors.New("unable to decode less than TLV"),
			SW_WRONG_DATA + 4:           errors.New("unable to decode greater than TLV"),
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_INCORRECT_P1P2:           errors.New("incorrect Parameters received"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_LENGTH:             errors.New("data length incorrect"),
			SW_WRONG_DATA:               errors.New("phonon index invalid"),
			SW_FILE_INVALID:             ErrInvalidPhononIndex,
			SW_FILE_INVALID + 1:         errors.New("phonon at index exceeds available phonon list"),
			SW_FILE_INVALID + 3:         errors.New("phonon at index is null"),
			SW_FILE_NOT_FOUND:           errors.New("phonon not initialized"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_LENGTH:             errors.New("incoming length wrong"),
			SW_WRONG_DATA:               errors.New("invalid phonon index"),
			SW_FILE_INVALID:             ErrInvalidPhononIndex,
			SW_FILE_INVALID + 1:         errors.New("phononon doesn't exist"),
			// adding 2 doesn't work because it conflicts with another error
			SW_FILE_INVALID + 3: errors.New("phonon already deleted"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_INCORRECT_P1P2:           errors.New("phononList continue greater than 1"),
			SW_INCORRECT_P1P2 + 1:       errors.New("no Phonons Requested"),
			SW_WRONG_DATA:               errors.New("incorrect phonon index"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: errors.New("phonon recipt conditions not met"),
			SW_FILE_FULL:                errors.New("maximum number of phonons exceeded"),
			SW_WRONG_DATA:               errors.New("unable to decode Phonon key list TLV"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_FILE_FULL:                errors.New("no phonon with index passed"),
			SW_WRONG_DATA:               errors.New("unable to decode Phonon key list TLV"),
			SW_WRONG_DATA + 1:           errors.New("unable to decode phonon key TLV"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_DATA:               errors.New("unable to decode TLV tag")},
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_DATA:               errors.New("unable to decode certificate TLV"),
			SW_COMMAND_NOT_ALLOWED:      errors.New("card certificate not initialized"),
		},
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_DATA:               errors.New("unable to decode card certificate TLV"),
			SW_WRONG_DATA + 1:           errors.New("unable to decode salt TLV"),
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
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
			SW_WRONG_DATA:               errors.New("unable to read salt"),
			SW_WRONG_DATA + 1:           errors.New("unable to read AES TLV"),
			SW_WRONG_DATA + 2:           errors.New("unable to read Signature TLV"),
		},
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
		PossibleErrs: CmdErrTable{
			// No idea how you can get this far without validating a pin
			SW_CONDITIONS_NOT_SATISFIED:      ErrPINNotEntered,
			SW_WRONG_DATA:                    errors.New("unable to read Receiver signature TLV"),
			SW_SECURITY_STATUS_NOT_SATISFIED: errors.New("unable to verify signature"),
		},
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
		PossibleErrs: CmdErrTable{
			SW_COMMAND_NOT_ALLOWED: errors.New("certificate already loaded"),
			SW_DATA_INVALID:        errors.New("unable to save certificate"),
		},
	}
}

// put here to be next to the select applet command function
var phononAID = []byte{0xA0, 0x00, 0x00, 0x08, 0x20, 0x00, 0x03, 0x01}

func NewCommandSelectPhononApplet() *Command {
	return &Command{
		ApduCmd: globalplatform.NewCommandSelect(phononAID),
		// no errors known
		PossibleErrs: CmdErrTable{},
	}
}

func NewCommandPairStep1(salt []byte, pairingPubKey *ecdsa.PublicKey) *Command {
	return &Command{
		ApduCmd: gridplus.NewAPDUPairStep1(salt, pairingPubKey),
		PossibleErrs: CmdErrTable{
			SW_WRONG_DATA:                     errors.New("data incorrect size"),
			SW_SECURE_MESSAGING_NOT_SUPPORTED: errors.New("no certificate loaded"),
			SW_SECURITY_STATUS_NOT_SATISFIED:  errors.New("unable to compute ECDH secrets"),
		},
	}

}

func NewCommandPairStep2(cryptogram [32]byte) *Command {
	return &Command{
		ApduCmd: gridplus.NewAPDUPairStep2(cryptogram[0:]),
		PossibleErrs: CmdErrTable{
			SW_WRONG_DATA:                    errors.New("wrong secret length"),
			SW_SECURITY_STATUS_NOT_SATISFIED: errors.New("client cryptogram differs from expected"),
		},
	}

}

func NewCommandUnpair(index uint8) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandUnpair(index),
		// No errors known
		PossibleErrs: CmdErrTable{},
	}

}

func NewCommandOpenSecureChannel(index uint8, publicKey []byte) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandOpenSecureChannel(index, publicKey),
		PossibleErrs: CmdErrTable{
			SW_INCORRECT_P1P2:                errors.New("incorrect parameters"),
			SW_SECURITY_STATUS_NOT_SATISFIED: errors.New("unable to generate secret"),
		},
	}

}

func NewCommandMutualAuthenticate(data []byte) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandMutuallyAuthenticate(data),
		PossibleErrs: CmdErrTable{
			SW_CONDITIONS_NOT_SATISFIED:      errors.New("authentication key not initialized"),
			SW_LOGICAL_CHANNEL_NOT_SUPPORTED: errors.New("already Mutually Authenticated"),
			SW_SECURITY_STATUS_NOT_SATISFIED: errors.New("secret length invalid"),
		},
	}

}

func NewCommandInit(data []byte) *Command {
	return &Command{
		ApduCmd: keycard.NewCommandInit(data),
		// No errors known
		PossibleErrs: CmdErrTable{},
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
		PossibleErrs: CmdErrTable{},
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
		PossibleErrs: CmdErrTable{},
	}
}

func NewCommandGetFriendlyName() *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsSetFriendlyName,
			0x00,
			0x00,
			[]byte{0x00},
		),
		//TODO: Errors
		PossibleErrs: CmdErrTable{
			SW_DATA_INVALID: errors.New("friendly name not set"),
		},
	}
}

func NewCommandSetFriendlyName(name string) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsGetFriendlyName,
			0x00,
			0x00,
			[]byte(name),
		),
		PossibleErrs: CmdErrTable{},
	}
}

func NewCommandGetAvailableMemory() *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsGetAvailableMemory,
			0x00,
			0x00,
			nil,
		),
	}
}

func NewCommandMineNativePhonon(difficulty uint8) *Command {
	return &Command{
		ApduCmd: apdu.NewCommand(
			globalplatform.ClaGp,
			InsMineNativePhonon,
			byte(difficulty),
			0x00,
			nil,
		),
		PossibleErrs: CmdErrTable{
			SW_MINING_FAILED:            ErrMiningFailed,
			SW_CONDITIONS_NOT_SATISFIED: ErrPINNotEntered,
		},
	}
}
