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
			KeyIndex:     int(binary.BigEndian.Uint16(keyIndexBytes)),
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

//checkPairingErrors takes a pairing step, either 1 or 2, and the SW value of the response to return appropriate error messages
func checkPairingErrors(pairingStep int, status uint16) (err error) {
	if pairingStep != 1 && pairingStep != 2 {
		return errors.New("pairing step must be set to 1 or 2 to check pairing errors")
	}
	switch status {
	case 0x6A80:
		err = errors.New("invalid pairing data format")
	case 0x6882:
		err = errors.New("certificate not loaded")
	case 0x6982:
		if pairingStep == 1 {
			err = errors.New("unable to generate secret")
		} else if pairingStep == 2 {
			err = errors.New("client cryptogram verification failed")
		}
	case 0x6A84:
		err = errors.New("all available pairing slots taken")
	case 0x6A86:
		err = errors.New("P1 invalid or first pairing phase was not completed")
	case 0x6985:
		err = errors.New("secure channel is already open")
	}

	return err
}
