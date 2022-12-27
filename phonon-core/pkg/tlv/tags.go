package tlv

const (
	// tags
	TagPairingSlots  = 0x03
	TagAppCapability = 0x8D

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
)
