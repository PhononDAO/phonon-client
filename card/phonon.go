package card

import (
	"encoding/binary"
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
