//go:generate stringer -type=CurrencyType,CurveType
package model

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/GridPlus/phonon-client/tlv"
	"github.com/GridPlus/phonon-client/util"
)

type Phonon struct {
	KeyIndex              uint16
	PubKey                *ecdsa.PublicKey
	CurveType             uint8
	SchemaVersion         uint8
	ExtendedSchemaVersion uint8
	Denomination          uint64
	CurrencyType          CurrencyType
	ExtendedTLV           []tlv.TLV
}

func (p *Phonon) String() string {
	return fmt.Sprintf("KeyIndex: %v, Denomination: %v, currencyType: %v, PubKey: %v\n",
		p.KeyIndex, p.Denomination, p.CurrencyType, util.ECDSAPubKeyToHexString(p.PubKey))
}

type CurrencyType uint16

const (
	Unspecified CurrencyType = 0x0000
	Bitcoin     CurrencyType = 0x0001
	Ethereum    CurrencyType = 0x0002
)

type CurveType uint8

const (
	Secp256k1 CurveType = iota
)
