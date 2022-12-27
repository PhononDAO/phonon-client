package smartcard

import (
	"encoding/binary"

	"github.com/GridPlus/phonon-core/pkg/model"
	"github.com/GridPlus/phonon-core/pkg/tlv"

	log "github.com/sirupsen/logrus"
)

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
