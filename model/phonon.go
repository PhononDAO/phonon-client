//go:generate stringer -type=CurrencyType,CurveType
package model

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/GridPlus/phonon-client/tlv"
	"github.com/GridPlus/phonon-client/util"
)

type Phonon struct {
	KeyIndex              uint16
	PubKey                *ecdsa.PublicKey
	CurveType             CurveType
	SchemaVersion         uint8
	ExtendedSchemaVersion uint8
	Denomination          Denomination
	CurrencyType          CurrencyType
	ExtendedTLV           []tlv.TLV
	Address               string //chain specific attribute not stored on card
	AddressType           uint8  //chain specific address type identifier
}

func (p *Phonon) String() string {
	return fmt.Sprintf("KeyIndex: %v\nDenomination: %v\ncurrencyType: %v\nPubKey: %v\nAddress: %v\nCurveType: %v\nSchemaVersion: %v\nExtendedSchemaVersion: %v\nExtendedTLV: %v\n",
		p.KeyIndex,
		p.Denomination,
		p.CurrencyType,
		util.ECCPubKeyToHexString(p.PubKey),
		p.Address,
		p.CurveType,
		p.SchemaVersion,
		p.ExtendedSchemaVersion,
		p.ExtendedTLV)
}

//Phonon data structured for display to the user an
type PhononJSON struct {
	KeyIndex              uint16
	PubKey                string //pubkey as hexstring
	Address               string //Chain specific address as hexstring
	AddressType           uint8
	SchemaVersion         uint8
	ExtendedSchemaVersion uint8
	Denomination          int
	CurrencyType          int
	//TODO extendedTLV
}

//Unmarshals a PhononUserView into an internal phonon representation
func (p *Phonon) UnmarshalJSON(b []byte) error {
	phuv := PhononJSON{}
	err := json.Unmarshal(b, &phuv)
	if err != nil {
		return err
	}
	p.KeyIndex = phuv.KeyIndex
	//Convert hexstring pubkey to *ecdsa.PublicKey
	pubKeyBytes, err := hex.DecodeString(phuv.PubKey)
	if err != nil {
		return err
	}
	p.PubKey, err = util.ParseECCPubKey(pubKeyBytes)
	if err != nil {
		return err
	}

	p.Address = phuv.Address
	p.AddressType = phuv.AddressType
	p.SchemaVersion = phuv.SchemaVersion
	p.ExtendedSchemaVersion = phuv.ExtendedSchemaVersion
	//Convert int to model.Denomination
	p.Denomination, err = NewDenomination(phuv.Denomination)
	if err != nil {
		return err
	}
	p.CurrencyType = CurrencyType(phuv.CurrencyType)

	return nil
}

func (p *Phonon) MarshalJSON() ([]byte, error) {
	userReqPhonon := &PhononJSON{
		KeyIndex:              p.KeyIndex,
		PubKey:                util.ECCPubKeyToHexString(p.PubKey),
		Address:               p.Address,
		SchemaVersion:         p.SchemaVersion,
		ExtendedSchemaVersion: p.ExtendedSchemaVersion,
		Denomination:          p.Denomination.Value(),
		CurrencyType:          int(p.CurrencyType),
		//TODO extendedTLV
	}
	jsonBytes, err := json.Marshal(userReqPhonon)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
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

type Denomination struct {
	Base     uint8
	Exponent uint8
}

//NewDenomination takes an integer input and attempts to store it as a compressible value representing currency base units
//Precision is limited to significant digits no greater than the value 255, along with exponentiation up to 255 digits
func NewDenomination(i int) (Denomination, error) {
	var exponent uint8
	//compress into exponent as much as possible
	for i > math.MaxUint8 {
		if i%10 == 0 {
			exponent += 1
			i = i / 10
		}
	}
	//If remaining base cannot be stored in a uint8 return error since this value can't be represented
	//Else return Denomination
	if i > math.MaxUint8 {
		return Denomination{}, errors.New("denomination exceeds representable precision")
	}
	return Denomination{
		Base:     uint8(i),
		Exponent: exponent,
	}, nil
}

func (d Denomination) Value() int {
	output := int(d.Base)
	exponent := d.Exponent
	for exponent != 0 {
		output *= 10
		exponent -= 1
	}
	return output
}

func (d Denomination) String() string {
	return fmt.Sprint(d.Value())
}
