package card

import (
	"crypto/ecdsa"
	"encoding/binary"
	"errors"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/tlv"
	"github.com/GridPlus/phonon-client/util"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

func encodeKeyIndexList(keyIndices []uint16) []byte {
	var keyIndexBytes []byte
	b := make([]byte, 2)
	for _, keyIndex := range keyIndices {
		binary.BigEndian.PutUint16(b, keyIndex)
		keyIndexBytes = append(keyIndexBytes, b...)
	}
	data, _ := tlv.NewTLV(TagPhononKeyIndexList, keyIndexBytes)
	return data.Encode()
}

func encodeSetDescriptorData(p *model.Phonon) ([]byte, error) {
	keyIndexBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(keyIndexBytes, p.KeyIndex)
	keyIndexTLV, err := tlv.NewTLV(TagKeyIndex, keyIndexBytes)
	if err != nil {
		return nil, err
	}

	phononTLV, err := TLVEncodePhononDescriptor(p)
	if err != nil {
		return nil, err
	}
	return append(keyIndexTLV.Encode(), phononTLV...), nil
}

func encodeListPhononsData(currencyType model.CurrencyType, lessThanValue uint64, greaterThanValue uint64) (p2 byte, data []byte, err error) {
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

	currencyTypeTLV, err := tlv.NewTLV(TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return p2, nil, err
	}
	//Translate filter values to bytes
	lessThanBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lessThanBytes, lessThanValue)

	greaterThanBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(greaterThanBytes, greaterThanValue)

	lessThanTLV, err := tlv.NewTLV(TagValueFilterLessThan, lessThanBytes)
	if err != nil {
		return p2, nil, err
	}
	greaterThanTLV, err := tlv.NewTLV(TagValueFilterMoreThan, greaterThanBytes)
	if err != nil {
		return p2, nil, err
	}

	innerData := tlv.EncodeTLVList(currencyTypeTLV, lessThanTLV, greaterThanTLV)
	cmdData, err := tlv.NewTLV(TagPhononFilter, innerData)
	if err != nil {
		return p2, nil, err
	}

	return p2, cmdData.Encode(), nil
}

func encodeSetReceiveListData(phononPubKeys []*ecdsa.PublicKey) ([]byte, error) {
	var pubKeyTLVBytes []byte
	for _, pubKey := range phononPubKeys {
		pubKeyTLV, _ := tlv.NewTLV(TagPhononPubKey, ethcrypto.FromECDSAPub(pubKey))
		pubKeyTLVBytes = append(pubKeyTLVBytes, pubKeyTLV.Encode()...)
	}

	data, err := tlv.NewTLV(TagPhononPubKeyList, pubKeyTLVBytes)
	if err != nil {
		return nil, err
	}
	return data.Encode(), nil
}

func encodeSendPhononsData(keyIndices []uint16) (data []byte, p2Length byte) {
	p2Length = byte(len(keyIndices))

	data = encodeKeyIndexList(keyIndices)
	return data, p2Length
}

func parseListPhononsResponse(resp []byte) ([]*model.Phonon, error) {
	phononCollection, err := tlv.ParseTLVPacket(resp, TagPhononCollection)
	if err != nil {
		return nil, err
	}
	//No phonons in list, the only tag will be the overall collection
	if len(phononCollection) <= 1 {
		return nil, nil
	}
	phonons := make([]*model.Phonon, 0)
	phononDescriptions, err := phononCollection.FindTags(TagPhononDescriptor)
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
	//Additionally parse KeyIndex
	keyIndexBytes, err := collection.FindTag(TagKeyIndex)
	if err != nil {
		return nil, err
	}
	phonon.KeyIndex = binary.BigEndian.Uint16(keyIndexBytes)
	return phonon, nil
}

func parseGetPhononPubKeyResponse(resp []byte) (pubKey *ecdsa.PublicKey, err error) {
	collection, err := tlv.ParseTLVPacket(resp, TagTransferPhononPacket)
	if err != nil {
		return nil, err
	}
	//Find interior phonon description tag
	description, err := collection.FindTag(TagPhononPrivateDescription)
	if err != nil {
		return nil, err
	}
	//Parse again to get TLV's nested under description
	descriptionTLV, err := tlv.ParseTLVPacket(description)
	if err != nil {
		return nil, err
	}

	rawPubKey, err := descriptionTLV.FindTag(TagPhononPubKey)
	if err != nil {
		return nil, err
	}
	pubKey, err = util.ParseECDSAPubKey(rawPubKey)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func parseDestroyPhononResponse(resp []byte) (privKey *ecdsa.PrivateKey, err error) {
	collection, err := tlv.ParseTLVPacket(resp)
	if err != nil {
		return nil, err
	}
	rawPrivKey, err := collection.FindTag(TagPhononPrivKey)
	if err != nil {
		return nil, err
	}
	privKey, err = util.ParseECCPrivKey(rawPrivKey)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func parseCreatePhononResponse(resp []byte) (keyIndex uint16, pubKey *ecdsa.PublicKey, err error) {
	collection, err := tlv.ParseTLVPacket(resp, TagPhononKeyCollection)
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

	keyIndex = binary.BigEndian.Uint16(keyIndexBytes)

	pubKey, err = util.ParseECDSAPubKey(pubKeyBytes)
	if err != nil {
		log.Error("could not parse pubkey from phonon response: ", err)
		return keyIndex, nil, err
	}

	return keyIndex, pubKey, nil
}

func parseSelectResponse(resp []byte) (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error) {
	if len(resp) == 0 {
		return nil, nil, false, errors.New("received nil response")
	}
	log.Debug("length of select response data: ", len(resp))
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
		cardPubKey, err = util.ParseECDSAPubKey(resp[22:87])
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
		cardPubKey, err = util.ParseECDSAPubKey(resp[2 : 2+length])
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
	cardPubKey, err = util.ParseECDSAPubKey(resp[2:67])
	if err != nil {
		return nil, nil, errors.New("could not parse card public key")
	}
	sig, err = util.ParseECDSASignature(resp[67:])
	if err != nil {
		return nil, nil, errors.New("could not parse card signature")
	}
	return cardPubKey, sig, nil
}

//Replacement for original parsing logic which did not retrieve full certificate
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

func parseGetAvailableMemoryResponse(data []byte) (persistentMem int, onResetMem int, onDeselectMem int, err error) {
	if len(data) != 12 {
		return 0, 0, 0, errors.New("response invalid length")
	}
	persistentMem = int(binary.BigEndian.Uint32(data[0:4]))
	onResetMem = int(binary.BigEndian.Uint32(data[4:8]))
	onDeselectMem = int(binary.BigEndian.Uint32(data[8:12]))

	return persistentMem, onResetMem, onDeselectMem, nil
}
