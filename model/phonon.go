//go:generate stringer -type=CurrencyType,CurveType
package model

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	log "github.com/sirupsen/logrus"

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
	ChainID               int //Not currently stored on card
	ExtendedTLV           []tlv.TLV
	Address               string //chain specific attribute not stored on card
	AddressType           uint8  //chain specific address type identifier
}

func (p *Phonon) String() string {
	return fmt.Sprintf("KeyIndex: %v\nDenomination: %v\nCurrencyType: %v\nPubKey: %v\nAddress: %v\nChainID: %v\nCurveType: %v\nSchemaVersion: %v\nExtendedSchemaVersion: %v\nExtendedTLV: %v\n",
		p.KeyIndex,
		p.Denomination,
		p.CurrencyType,
		util.ECCPubKeyToHexString(p.PubKey),
		p.Address,
		p.ChainID,
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
	Denomination          Denomination
	CurrencyType          int
	ChainID               int
}

//Unmarshals a PhononUserView into an internal phonon representation
func (p *Phonon) UnmarshalJSON(b []byte) error {
	phJSON := PhononJSON{}
	err := json.Unmarshal(b, &phJSON)
	if err != nil {
		return err
	}
	p.KeyIndex = phJSON.KeyIndex
	//Convert hexstring pubkey to *ecdsa.PublicKey
	pubKeyBytes, err := hex.DecodeString(phJSON.PubKey)
	if err != nil {
		return err
	}
	p.PubKey, err = util.ParseECCPubKey(pubKeyBytes)
	if err != nil {
		return err
	}

	p.Address = phJSON.Address
	p.AddressType = phJSON.AddressType
	p.SchemaVersion = phJSON.SchemaVersion
	p.ExtendedSchemaVersion = phJSON.ExtendedSchemaVersion
	p.Denomination = phJSON.Denomination
	p.CurrencyType = CurrencyType(phJSON.CurrencyType)
	p.ChainID = phJSON.ChainID

	return nil
}

func (p *Phonon) MarshalJSON() ([]byte, error) {
	userReqPhonon := &PhononJSON{
		KeyIndex:              p.KeyIndex,
		PubKey:                util.ECCPubKeyToHexString(p.PubKey),
		Address:               p.Address,
		SchemaVersion:         p.SchemaVersion,
		ExtendedSchemaVersion: p.ExtendedSchemaVersion,
		Denomination:          p.Denomination,
		CurrencyType:          int(p.CurrencyType),
		ChainID:               p.ChainID,
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

var ErrInvalidDenomination = errors.New("value cannot be represented as a phonon denomination")

//NewDenomination takes an integer input and attempts to store it as a compressible value representing currency base units
//Precision is limited to significant digits no greater than the value 255, along with exponentiation up to 255 digits
func NewDenomination(i *big.Int) (Denomination, error) {
	var exponent uint8
	maxUint8 := big.NewInt(int64(math.MaxUint8))
	//compress into exponent as much as possible
	var m *big.Int
	for i.Cmp(maxUint8) > 0 {
		m = new(big.Int).Mod(i, big.NewInt(10))
		log.Debug("m = ", m)
		if m.Cmp(big.NewInt(0)) == 0 {
			exponent += 1
			i.Div(i, big.NewInt(10))
			log.Debugf("i = %v, exponent = %v", i, exponent)
		} else {
			return Denomination{}, ErrInvalidDenomination
		}
	}
	//If remaining base cannot be stored in a uint8 return error since this value can't be represented
	//Else return Denomination
	if i.Cmp(maxUint8) > 0 {
		log.Error("remaining denomination base = ", i)
		return Denomination{}, errors.New("denomination exceeds representable precision")
	}
	log.Debugf("i = %v, exponent = %v", i, exponent)
	d := Denomination{
		Base:     uint8(i.Int64()),
		Exponent: exponent,
	}
	log.Debugf("d: %+v", d)
	return d, nil
}

func (d Denomination) Value() *big.Int {
	//TODO: Convert to *big.Int
	output := big.NewInt(int64(d.Base))
	exponent := d.Exponent
	for exponent != 0 {
		output.Mul(output, big.NewInt(10))
		exponent -= 1
	}
	return output
}

func (d Denomination) String() string {
	return fmt.Sprint(d.Value())
}

//Unmarshal Denominations from string to internal representation
func (d *Denomination) UnmarshalJSON(b []byte) error {
	var denomString string
	err := json.Unmarshal(b, &denomString)
	if err != nil {
		return err
	}
	denomBigInt, ok := big.NewInt(0).SetString(denomString, 10)
	if !ok {
		return errors.New("denomination string not representable as *big.Int")
	}
	denom, err := NewDenomination(denomBigInt)
	if err != nil {
		return err
	}
	//d must refer to itself to mutate its value
	*d = denom
	return nil
}

//Marshal denominations as strings
func (d *Denomination) MarshalJSON() ([]byte, error) {
	dString := d.String()
	data, err := json.Marshal(dString)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//TODO: Denomination JSON Marshalling
