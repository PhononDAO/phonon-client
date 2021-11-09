package card

import (
	"encoding/binary"
	"errors"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/tlv"

	log "github.com/sirupsen/logrus"
)

//TLV Encodes the phonon standard schema used for setting it's descriptor. Must be extended with additional fields
//to suit the various commands that deal with phonons.
//Excludes fields KeyIndex, PubKey, and CurveType
//Includes fields SchemaVersion, ExtendedSchemaVersion, CurrencyType, Denomination, and ExtendedTLVs
func TLVEncodePhononDescriptor(p *model.Phonon) ([]byte, error) {
	//KeyIndex omitted

	//PubKey omitted

	//CurveType omitted

	log.Debug("encoding phonon: ", p)
	schemaVersionTLV, err := tlv.NewTLV(TagSchemaVersion, []byte{p.SchemaVersion})
	if err != nil {
		return nil, err
	}
	extendedSchemaVersionTLV, err := tlv.NewTLV(TagExtendedSchemaVersion, []byte{p.SchemaVersion})
	if err != nil {
		return nil, err
	}

	// denominationBytes := make([]byte, 8)
	// binary.BigEndian.PutUint64(denominationBytes, p.Denomination)

	denomBaseTLV, err := tlv.NewTLV(TagPhononDenomBase, []byte{byte(p.Denomination.Base)})
	if err != nil {
		return nil, err
	}
	denomExpTLV, err := tlv.NewTLV(TagPhononDenomExp, []byte{byte(p.Denomination.Exponent)})
	if err != nil {
		return nil, err
	}

	currencyTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyTypeBytes, uint16(p.CurrencyType))
	currencyTypeTLV, err := tlv.NewTLV(TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return nil, err
	}

	phononTLV := append(schemaVersionTLV.Encode(), extendedSchemaVersionTLV.Encode()...)
	phononTLV = append(phononTLV, denomBaseTLV.Encode()...)
	phononTLV = append(phononTLV, denomExpTLV.Encode()...)
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
	phonon.CurveType = model.CurveType(rawCurveType[0])

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
	denomBaseBytes, err := phononTLV.FindTag(TagPhononDenomBase)
	if err != nil {
		log.Debug("could not parse denomination base: ", err)
		return phonon, err
	}
	if len(denomBaseBytes) != 1 {
		return phonon, errors.New("denomBaseBytes length incorrect")
	}
	phonon.Denomination.Base = uint8(denomBaseBytes[0])

	denomExpBytes, err := phononTLV.FindTag(TagPhononDenomExp)
	if err != nil {
		log.Debug("could not parse denomination exponent: ", err)
		return phonon, err
	}
	if len(denomExpBytes) != 1 {
		return phonon, errors.New("denomBaseExp length incorrect")
	}
	phonon.Denomination.Exponent = uint8(denomExpBytes[0])

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
		TagPhononDenomBase, TagPhononDenomExp, TagCurrencyType}
	phonon.ExtendedTLV = phononTLV.GetRemainingTLVs(standardTags)

	return phonon, nil
}
