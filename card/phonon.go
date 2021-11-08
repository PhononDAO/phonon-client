package card

import (
	"encoding/binary"
	"errors"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/tlv"

	log "github.com/sirupsen/logrus"
)

//TLV Encodes the phonon standard schema, excepting fields KeyIndex and PubKey
func TLVEncodeStandardPhonon(p *model.Phonon) ([]byte, error) {
	//KeyIndex omitted

	//PubKey omitted

	log.Debug("encoding phonon: ", p)
	curveTypeTLV, err := tlv.NewTLV(TagCurveType, []byte{p.CurveType})
	if err != nil {
		return nil, err
	}
	schemaVersionTLV, err := tlv.NewTLV(TagSchemaVersion, []byte{p.SchemaVersion})
	if err != nil {
		return nil, err
	}
	extendedSchemaVersionTLV, err := tlv.NewTLV(TagExtendedSchemaVersion, []byte{p.SchemaVersion})
	if err != nil {
		return nil, err
	}

	denominationBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(denominationBytes, p.Denomination)

	valueTLV, err := tlv.NewTLV(TagPhononValue, denominationBytes)
	if err != nil {
		return nil, err
	}

	currencyTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyTypeBytes, uint16(p.CurrencyType))
	currencyTypeTLV, err := tlv.NewTLV(TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return nil, err
	}

	phononTLV := append(curveTypeTLV.Encode(), schemaVersionTLV.Encode()...)
	phononTLV = append(phononTLV, extendedSchemaVersionTLV.Encode()...)
	phononTLV = append(phononTLV, valueTLV.Encode()...)
	phononTLV = append(phononTLV, currencyTypeTLV.Encode()...)
	for _, field := range p.ExtendedTLV {
		phononTLV = append(phononTLV, field.Encode()...)
	}

	return phononTLV, nil
}

//Decodes the public phonon fields typically returned from a card
//Excludes PubKey and KeyIndex
func TLVDecodePublicPhononFields(phononTLV tlv.TLVCollection) (*model.Phonon, error) {
	phonon := &model.Phonon{}
	for tag, _ := range phononTLV {
		log.Debugf("found tag in decoding: % X", tag)
	}
	//CurveType
	//TODO: inspect supported curve types
	rawCurveType, err := phononTLV.FindTag(TagCurveType)
	if err != nil {
		return nil, err
	}
	if len(rawCurveType) != 1 {
		return nil, errors.New("curveType length incorrect")
	}
	phonon.CurveType = uint8(rawCurveType[0])

	//SchemaVersion
	rawSchemaVersion, err := phononTLV.FindTag(TagSchemaVersion)
	if err != nil {
		log.Debug("could not parse schema version tag")
		return phonon, err
	}
	log.Debug("value of rawSchemaVersion: ", rawSchemaVersion)
	if len(rawSchemaVersion) != 1 {
		return phonon, errors.New("schemaVersion length incorrect")
	}
	phonon.SchemaVersion = uint8(rawSchemaVersion[0])

	if phonon.SchemaVersion != StandardSchemaSupportedVersions {
		return phonon, errors.New("unsupported phonon standard schema version")
	}

	rawExtendedSchemaVersion, err := phononTLV.FindTag(TagExtendedSchemaVersion)
	if err != nil {
		log.Debug("could not parse extended schema version tag")
		return phonon, err
	}
	if len(rawExtendedSchemaVersion) != 1 {
		return phonon, errors.New("extendedSchemaVersion length incorrect")
	}
	phonon.ExtendedSchemaVersion = uint8(rawExtendedSchemaVersion[0])

	//Denomination
	denominationBytes, err := phononTLV.FindTag(TagPhononValue)
	if err != nil {
		log.Debug("could not parse denomination tag")
		return phonon, err
	}
	phonon.Denomination = binary.BigEndian.Uint64(denominationBytes)

	//CurrencyType
	currencyTypeBytes, err := phononTLV.FindTag(TagCurrencyType)
	if err != nil {
		log.Debug("could not parse currencyType tag")
		return phonon, err
	}
	log.Debugf("currencyTypeBytes: % X", currencyTypeBytes)
	phonon.CurrencyType = model.CurrencyType(binary.BigEndian.Uint16(currencyTypeBytes))

	//Extended Schema

	//Standard Schema Tags
	standardTags := []byte{TagPhononPrivKey, TagCurveType, TagSchemaVersion, TagExtendedSchemaVersion,
		TagPhononValue, TagCurrencyType}
	phonon.ExtendedTLV = phononTLV.GetRemainingTLVs(standardTags)

	return phonon, nil
}
