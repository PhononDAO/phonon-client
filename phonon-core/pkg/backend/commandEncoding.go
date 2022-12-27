package backend

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"

	"github.com/GridPlus/phonon-core/internal/util"
	"github.com/GridPlus/phonon-core/pkg/cert"
	"github.com/GridPlus/phonon-core/pkg/model"
	"github.com/GridPlus/phonon-core/pkg/tlv"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	log "github.com/sirupsen/logrus"
)

const StandardSchemaSupportedVersions uint8 = 0

var (
	ErrMiningFailed       = errors.New("native phonon mine attempt failed")
	ErrInvalidPhononIndex = errors.New("invalid phonon index")
	ErrInvalidKeyLength   = errors.New("key invalid length")
	ErrCertLocked         = errors.New("certificate already locked. cannot be reset")
	ErrDefault            = errors.New("unspecified error for command")
)

func EncodeSetReceiveListData(phononPubKeys []*ecdsa.PublicKey) ([]byte, error) {
	var pubKeyTLVBytes []byte
	for _, pubKey := range phononPubKeys {
		pubKeyTLV, _ := tlv.NewTLV(tlv.TagPhononPubKey, ethcrypto.FromECDSAPub(pubKey))
		pubKeyTLVBytes = append(pubKeyTLVBytes, pubKeyTLV.Encode()...)
	}

	data, err := tlv.NewTLV(tlv.TagPhononPubKeyList, pubKeyTLVBytes)
	if err != nil {
		return nil, err
	}
	return data.Encode(), nil
}

func EncodeKeyIndexList(keyIndices []model.PhononKeyIndex) []byte {
	var keyIndexBytes []byte
	for _, keyIndex := range keyIndices {
		b := keyIndex.ToBytes()
		keyIndexBytes = append(keyIndexBytes, b...)
	}
	data, _ := tlv.NewTLV(tlv.TagPhononKeyIndexList, keyIndexBytes)
	return data.Encode()
}

func ParseListPhononsResponse(resp []byte) ([]*model.Phonon, error) {
	phononCollection, err := tlv.ParseTLVPacket(resp, tlv.TagPhononCollection)
	if err != nil {
		return nil, err
	}
	//No phonons in list, the only tag will be the overall collection
	if len(phononCollection) <= 1 {
		return nil, nil
	}
	phonons := make([]*model.Phonon, 0)
	phononDescriptions, err := phononCollection.FindTags(tlv.TagPhononDescriptor)
	if err != nil {
		return nil, err
	}

	for _, description := range phononDescriptions {
		phonon, err := ParsePhononDescriptor(description)
		if err != nil {
			log.Error("unable to parse phonon: ", err)
		}

		phonons = append(phonons, phonon)
	}
	return phonons, nil
}

func ParsePhononDescriptor(description []byte) (*model.Phonon, error) {
	collection, err := tlv.ParseTLVPacket(description)
	if err != nil {
		return nil, err
	}
	phonon, err := TLVDecodePublicPhononFields(collection)
	if err != nil {
		return nil, err
	}
	return phonon, nil
}

func ParseGetPhononPubKeyResponse(resp []byte) (rawPubKey []byte, err error) {
	collection, err := tlv.ParseTLVPacket(resp, tlv.TagTransferPhononPacket)
	if err != nil {
		return nil, err
	}
	//Find interior phonon description tag
	description, err := collection.FindTag(tlv.TagPhononPrivateDescription)
	if err != nil {
		return nil, err
	}
	//Parse again to get TLV's nested under description
	descriptionTLV, err := tlv.ParseTLVPacket(description)
	if err != nil {
		return nil, err
	}

	rawPubKey, err = descriptionTLV.FindTag(tlv.TagPhononPubKey)
	if err != nil {
		return nil, err
	}
	return rawPubKey, nil
}

func ParseDestroyPhononResponse(resp []byte) (privKey *ecdsa.PrivateKey, err error) {
	collection, err := tlv.ParseTLVPacket(resp)
	if err != nil {
		return nil, err
	}
	rawPrivKey, err := collection.FindTag(tlv.TagPhononPrivKey)
	if err != nil {
		return nil, err
	}

	privKey, err = util.ParseECCPrivKey(rawPrivKey)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func ParseCreatePhononResponse(resp []byte) (keyIndex model.PhononKeyIndex, pubKeyBytes []byte, err error) {
	collection, err := tlv.ParseTLVPacket(resp, tlv.TagPhononKeyCollection)
	if err != nil {
		return 0, nil, err
	}

	keyIndexBytes, err := collection.FindTag(tlv.TagKeyIndex)
	if err != nil {
		return 0, nil, err
	}

	pubKeyBytes, err = collection.FindTag(tlv.TagPhononPubKey)
	if err != nil {
		return 0, nil, err
	}

	keyIndex = model.KeyIndexFromBytes(keyIndexBytes)

	return keyIndex, pubKeyBytes, nil
}

func ParseSelectResponse(resp []byte) (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error) {
	if len(resp) == 0 {
		return nil, nil, false, errors.New("received nil response")
	}
	log.Trace("length of select response data: ", len(resp))
	switch resp[0] {
	//Initialized
	case 0xA4:
		log.Debug("pin initialized")
		//If length of length is set this is a long format TLV response
		if len(resp) < 88 {
			log.Error("response should have been at least length 86 bytes, was length: ", len(resp))
			return nil, nil, false, errors.New("invalid response length")
		}
		instanceUID = resp[4:20]
		cardPubKey, err = util.ParseECCPubKey(resp[22:87])
		if err != nil {
			log.Error("select could not parse returned public key")
			return nil, nil, false, err
		}
		cardInitialized = true
		//Think this response pattern only existed when a safecard wallet was initialized
		// if resp[3] == 0x81 {
		// 	instanceUID = resp[6:22]
		// 	cardPubKey = resp[24:89]
		// } else {
		//This would be the existing parsing above
	case 0x80:
		log.Debug("pin uninitialized")
		length := int(resp[1])
		cardPubKey, err = util.ParseECCPubKey(resp[2 : 2+length])
		if err != nil {
			log.Error("select could not parse returned public key")
			return nil, nil, false, err
		}
		cardInitialized = false
		return nil, cardPubKey, false, nil
	}

	return instanceUID, cardPubKey, cardInitialized, nil
}

func ParseIdentifyCardResponse(resp []byte) (cardPubKey *ecdsa.PublicKey, sig *util.ECDSASignature, err error) {
	correctLength := 67
	if len(resp) < correctLength {
		log.Errorf("identify card response invalid length %v should be %v ", len(resp), correctLength)
		return nil, nil, err
	}
	cardPubKey, err = util.ParseECCPubKey(resp[2:67])
	if err != nil {
		return nil, nil, errors.New("could not parse card public key")
	}
	sig, err = util.ParseECDSASignature(resp[67:])
	if err != nil {
		return nil, nil, errors.New("could not parse card signature")
	}
	return cardPubKey, sig, nil
}

// Replacement for original parsing logic which did not retrieve full certificate
func ParsePairStep1Response(resp []byte) (salt []byte, cardCert cert.CardCertificate, pairingSig []byte, err error) {
	if len(resp) < 34 {
		return nil, cardCert, nil, errors.New("pairing response was invalid length")
	}
	salt = make([]byte, 32)
	copy(salt, resp[0:32])

	certLength := int(resp[33])

	if len(resp) < 43+certLength {
		return nil, cardCert, nil, errors.New("pairing response was invalid length")
	}
	rawCert := make([]byte, len(resp[32:34+certLength]))
	copy(rawCert, resp[32:34+certLength])
	cardCert, err = cert.ParseRawCardCertificate(rawCert)
	if err != nil {
		return nil, cardCert, nil, err
	}

	log.Debugf("end of resp len(%v): % X", len(resp[34+certLength:]), resp[34+certLength:])
	pairingSig = make([]byte, len(resp[34+certLength:]))
	copy(pairingSig, resp[34+certLength:])

	return salt, cardCert, pairingSig, nil
}

func ParseGetAvailableMemoryResponse(resp []byte) (persistentMem int, onResetMem int, onDeselectMem int, err error) {
	if len(resp) != 12 {
		return 0, 0, 0, errors.New("response invalid length")
	}
	persistentMem = int(binary.BigEndian.Uint32(resp[0:4]))
	onResetMem = int(binary.BigEndian.Uint32(resp[4:8]))
	onDeselectMem = int(binary.BigEndian.Uint32(resp[8:12]))

	return persistentMem, onResetMem, onDeselectMem, nil
}

func ParseMineNativePhononResponse(resp []byte) (keyIndex model.PhononKeyIndex, hash []byte, err error) {
	collection, err := tlv.ParseTLVPacket(resp, tlv.TagPhononKeyCollection)
	if err != nil {
		return 0, nil, err
	}
	keyIndexBytes, err := collection.FindTag(tlv.TagKeyIndex)
	if err != nil {
		return 0, nil, err
	}
	keyIndex = model.KeyIndexFromBytes(keyIndexBytes)

	hash, err = collection.FindTag(tlv.TagPhononPubKey)
	if err != nil {
		return 0, nil, err
	}
	return keyIndex, hash, nil
}

// Decodes the public phonon fields typically returned from a card
// Excludes PubKey and KeyIndex
func TLVDecodePublicPhononFields(phononTLV tlv.TLVCollection) (*model.Phonon, error) {
	phonon := &model.Phonon{}

	//Optionally parse KeyIndex, present in ListPhonons but not during mock's ReceivePhonons
	keyIndexBytes, err := phononTLV.FindTag(tlv.TagKeyIndex)
	if err != nil && err != tlv.ErrTagNotFound {
		return nil, err
	}
	switch err {
	case nil:
		phonon.KeyIndex = model.KeyIndexFromBytes(keyIndexBytes)
	case tlv.ErrTagNotFound:
		log.Debug("phonon keyIndex not found during tlv parsing, skipping...")
	default:
		return nil, err

	}
	//CurveType
	rawCurveType, err := phononTLV.FindTag(tlv.TagCurveType)
	if err != nil {
		return nil, err
	}
	if len(rawCurveType) != 1 {
		return nil, errors.New("curveType length incorrect")
	}
	phonon.CurveType = model.CurveType(rawCurveType[0])

	//SchemaVersion
	rawSchemaVersion, err := phononTLV.FindTag(tlv.TagSchemaVersion)
	if err != nil {
		log.Debug("could not parse schema version tag")
		return phonon, err
	}
	if len(rawSchemaVersion) != 1 {
		return phonon, errors.New("schemaVersion length incorrect")
	}
	phonon.SchemaVersion = uint8(rawSchemaVersion[0])

	if phonon.SchemaVersion != StandardSchemaSupportedVersions {
		return phonon, errors.New("unsupported phonon standard schema version")
	}

	rawExtendedSchemaVersion, err := phononTLV.FindTag(tlv.TagExtendedSchemaVersion)
	if err != nil {
		log.Debug("could not parse extended schema version tag")
		return phonon, err
	}
	if len(rawExtendedSchemaVersion) != 1 {
		return phonon, errors.New("extendedSchemaVersion length incorrect")
	}
	phonon.ExtendedSchemaVersion = uint8(rawExtendedSchemaVersion[0])

	//Denomination
	denomBaseBytes, err := phononTLV.FindTag(tlv.TagPhononDenomBase)
	if err != nil {
		log.Debug("could not parse denomination base: ", err)
		return phonon, err
	}
	if len(denomBaseBytes) != 1 {
		return phonon, errors.New("denomBaseBytes length incorrect")
	}
	phonon.Denomination.Base = uint8(denomBaseBytes[0])

	denomExpBytes, err := phononTLV.FindTag(tlv.TagPhononDenomExp)
	if err != nil {
		log.Debug("could not parse denomination exponent: ", err)
		return phonon, err
	}
	if len(denomExpBytes) != 1 {
		return phonon, errors.New("denomBaseExp length incorrect")
	}
	phonon.Denomination.Exponent = uint8(denomExpBytes[0])

	//CurrencyType
	currencyTypeBytes, err := phononTLV.FindTag(tlv.TagCurrencyType)
	if err != nil {
		log.Debug("could not parse currencyType tag")
		return phonon, err
	}
	phonon.CurrencyType = model.CurrencyType(binary.BigEndian.Uint16(currencyTypeBytes))

	//Extended Schema

	//Standard Schema Tags
	standardTags := []byte{tlv.TagKeyIndex, tlv.TagPhononPrivKey, tlv.TagCurveType, tlv.TagSchemaVersion, tlv.TagExtendedSchemaVersion,
		tlv.TagPhononDenomBase, tlv.TagPhononDenomExp, tlv.TagCurrencyType}
	phonon.ExtendedTLV = phononTLV.GetRemainingTLVs(standardTags)

	//Collecting ChainID from extended tags pending a more elegant way to do this
	for _, entry := range phonon.ExtendedTLV {
		if entry.Tag == tlv.TagChainID {
			//guard parsing against panics
			if len(entry.Value) == 1 {
				phonon.ChainID = int(entry.Value[0])
			}
		}
	}
	return phonon, nil
}

func EncodeSetDescriptorData(p *model.Phonon) ([]byte, error) {
	keyIndexBytes := p.KeyIndex.ToBytes()
	keyIndexTLV, err := tlv.NewTLV(tlv.TagKeyIndex, keyIndexBytes)
	if err != nil {
		return nil, err
	}

	phononTLV, err := TLVEncodePhononDescriptor(p)
	if err != nil {
		return nil, err
	}
	return append(keyIndexTLV.Encode(), phononTLV...), nil
}

// TLV Encodes the phonon standard schema used for setting it's descriptor. Must be extended with additional fields
// to suit the various commands that deal with phonons.
// Excludes fields KeyIndex, PubKey, and CurveType which are already known by the card at the time of creation
// Includes fields SchemaVersion, ExtendedSchemaVersion, CurrencyType, Denomination, and ExtendedTLVs
func TLVEncodePhononDescriptor(p *model.Phonon) ([]byte, error) {
	//KeyIndex omitted

	//PubKey omitted

	//CurveType omitted

	log.Debug("encoding phonon: ", p)
	schemaVersionTLV, err := tlv.NewTLV(tlv.TagSchemaVersion, []byte{p.SchemaVersion})
	if err != nil {
		return nil, err
	}
	extendedSchemaVersionTLV, err := tlv.NewTLV(tlv.TagExtendedSchemaVersion, []byte{p.SchemaVersion})
	if err != nil {
		return nil, err
	}

	// denominationBytes := make([]byte, 8)
	// binary.BigEndian.PutUint64(denominationBytes, p.Denomination)

	denomBaseTLV, err := tlv.NewTLV(tlv.TagPhononDenomBase, []byte{byte(p.Denomination.Base)})
	if err != nil {
		return nil, err
	}
	denomExpTLV, err := tlv.NewTLV(tlv.TagPhononDenomExp, []byte{byte(p.Denomination.Exponent)})
	if err != nil {
		return nil, err
	}

	currencyTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyTypeBytes, uint16(p.CurrencyType))
	currencyTypeTLV, err := tlv.NewTLV(tlv.TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return nil, err
	}
	//adding chainID support through extended schema to avoid card code change
	chainIDTLV, err := tlv.NewTLV(tlv.TagChainID, []byte{byte(p.ChainID)})
	if err != nil {
		return nil, err
	}
	p.ExtendedTLV = append(p.ExtendedTLV, chainIDTLV)

	phononTLV := append(schemaVersionTLV.Encode(), extendedSchemaVersionTLV.Encode()...)
	phononTLV = append(phononTLV, denomBaseTLV.Encode()...)
	phononTLV = append(phononTLV, denomExpTLV.Encode()...)
	phononTLV = append(phononTLV, currencyTypeTLV.Encode()...)
	for _, field := range p.ExtendedTLV {
		phononTLV = append(phononTLV, field.Encode()...)
	}

	return phononTLV, nil
}

func EncodeListPhononsData(currencyType model.CurrencyType, lessThanValue uint64, greaterThanValue uint64) (p2 byte, data []byte, err error) {
	//Toggle filter bytes for nonzero lessThan and greaterThan filter values
	if lessThanValue == 0 {
		//Don't filter on value at all
		if greaterThanValue == 0 {
			p2 = 0x00
		}
		//Filter on only GreaterThan Value
		if greaterThanValue > 0 {
			p2 = 0x02
		}
	}
	if lessThanValue > 0 {
		//Filter on only LessThanValue
		if greaterThanValue == 0 {
			p2 = 0x01
		}
		//Filter on LessThan and GreaterThan
		if greaterThanValue > 0 {
			p2 = 0x03
		}

	}

	//Translate currencyType to bytes
	currencyTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyTypeBytes, uint16(currencyType))

	currencyTypeTLV, err := tlv.NewTLV(tlv.TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return p2, nil, err
	}
	//Translate filter values to bytes
	lessThanBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lessThanBytes, lessThanValue)

	greaterThanBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(greaterThanBytes, greaterThanValue)

	lessThanTLV, err := tlv.NewTLV(tlv.TagValueFilterLessThan, lessThanBytes)
	if err != nil {
		return p2, nil, err
	}
	greaterThanTLV, err := tlv.NewTLV(tlv.TagValueFilterMoreThan, greaterThanBytes)
	if err != nil {
		return p2, nil, err
	}

	innerData := tlv.EncodeTLVList(currencyTypeTLV, lessThanTLV, greaterThanTLV)
	cmdData, err := tlv.NewTLV(tlv.TagPhononFilter, innerData)
	if err != nil {
		return p2, nil, err
	}

	return p2, cmdData.Encode(), nil
}

func EncodeSendPhononsData(keyIndices []model.PhononKeyIndex) (data []byte, p2Length byte) {
	p2Length = byte(len(keyIndices))

	data = EncodeKeyIndexList(keyIndices)
	return data, p2Length
}
