package card

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	log "github.com/sirupsen/logrus"
)

func encodeSetDescriptorData(keyIndex uint16, currencyType model.CurrencyType, value float32) ([]byte, error) {
	keyIndexBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(keyIndexBytes, keyIndex)
	keyIndexTLV, err := NewTLV(TagKeyIndex, keyIndexBytes)
	if err != nil {
		return nil, err
	}

	currencyBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyBytes, uint16(currencyType))
	currencyTypeTLV, err := NewTLV(TagCurrencyType, currencyBytes)
	if err != nil {
		return nil, err
	}

	var valueBytes bytes.Buffer
	err = binary.Write(&valueBytes, binary.BigEndian, value)
	if err != nil {
		log.Error("unable to write float value as bytes: ", err)
		return nil, err
	}
	valueTLV, err := NewTLV(TagPhononValue, valueBytes.Bytes())
	if err != nil {
		return nil, err
	}

	descriptorBytes := append(keyIndexTLV.Encode(), currencyTypeTLV.Encode()...)
	descriptorBytes = append(descriptorBytes, valueTLV.Encode()...)
	phononDescriptorTLV, err := NewTLV(TagPhononDescriptor, descriptorBytes)
	if err != nil {
		return nil, err
	}
	return phononDescriptorTLV.Encode(), nil
}

func encodeListPhononsData(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) (p2 byte, data []byte, err error) {
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

	currencyTypeTLV, err := NewTLV(TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return p2, nil, err
	}
	//Translate filter values to bytes
	lessThanBytes, err := util.Float32ToBytes(lessThanValue)
	if err != nil {
		return p2, nil, err
	}
	greaterThanBytes, err := util.Float32ToBytes(greaterThanValue)
	if err != nil {
		return p2, nil, err
	}
	lessThanTLV, err := NewTLV(TagValueFilterLessThan, lessThanBytes)
	if err != nil {
		return p2, nil, err
	}
	greaterThanTLV, err := NewTLV(TagValueFilterMoreThan, greaterThanBytes)
	if err != nil {
		return p2, nil, err
	}

	innerData := EncodeTLVList(currencyTypeTLV, lessThanTLV, greaterThanTLV)
	cmdData, err := NewTLV(TagPhononFilter, innerData)
	if err != nil {
		return p2, nil, err
	}

	return p2, cmdData.Encode(), nil
}

func parseListPhononsResponse(resp []byte) ([]model.Phonon, error) {
	phononCollection, err := ParseTLVPacket(resp, TagPhononCollection)
	if err != nil {
		return nil, err
	}
	//No phonons in list, the only tag will be the overall collection
	if len(phononCollection) <= 1 {
		return nil, nil
	}
	phonons := make([]model.Phonon, 0)
	phononDescriptions, err := phononCollection.FindTags(TagPhononDescriptor)
	if err != nil {
		return nil, err
	}

	for _, description := range phononDescriptions {
		descriptionTLV, err := ParseTLVPacket(description)
		if err != nil {
			return phonons, err
		}
		keyIndexBytes, err := descriptionTLV.FindTag(TagKeyIndex)
		if err != nil {
			return phonons, err
		}
		currencyTypeBytes, err := descriptionTLV.FindTag(TagCurrencyType)
		if err != nil {
			return phonons, err
		}
		currencyType := binary.BigEndian.Uint16(currencyTypeBytes)

		valueBytes, err := descriptionTLV.FindTag(TagPhononValue)
		if err != nil {
			return phonons, err
		}

		var value float32
		err = binary.Read(bytes.NewReader(valueBytes), binary.BigEndian, &value)
		if err != nil {
			return phonons, err
		}
		phonon := model.Phonon{
			KeyIndex:     binary.BigEndian.Uint16(keyIndexBytes),
			CurrencyType: model.CurrencyType(currencyType),
			Value:        value,
		}
		phonons = append(phonons, phonon)
	}
	return phonons, nil
}

func parseGetPhononPubKeyResponse(resp []byte) (pubKey *ecdsa.PublicKey, err error) {
	collection, err := ParseTLVPacket(resp)
	if err != nil {
		return nil, err
	}
	rawPubKey, err := collection.FindTag(TagPhononPubKey)
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
	collection, err := ParseTLVPacket(resp)
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

	keyIndex = binary.BigEndian.Uint16(keyIndexBytes)

	pubKey, err = util.ParseECDSAPubKey(pubKeyBytes)
	if err != nil {
		log.Error("could not parse pubkey from phonon response: ", err)
		return keyIndex, nil, err
	}

	return keyIndex, pubKey, nil
}

func parseSelectResponse(resp []byte) (instanceUID []byte, cardPubKey []byte, err error) {
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

func parseIdentifyCardResponse(resp []byte) (cardPubKey []byte, sig []byte, err error) {
	correctLength := 67
	if len(resp) < 67 {
		log.Errorf("identify card response invalid length %v should be %v ", len(resp), correctLength)
		return nil, nil, err
	}
	cardPubKey = resp[2:67]
	sig = resp[67:]

	return cardPubKey, sig, nil
}
