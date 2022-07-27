//go:generate stringer -type=CurrencyType,CurveType
package model

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/GridPlus/phonon-client/tlv"
	"github.com/GridPlus/phonon-client/util"
	"github.com/ethereum/go-ethereum/crypto"

	log "github.com/sirupsen/logrus"
)

type Phonon struct {
	KeyIndex              PhononKeyIndex
	PubKey                PhononPubKey
	CurveType             CurveType
	SchemaVersion         uint8
	ExtendedSchemaVersion uint8
	Denomination          Denomination
	CurrencyType          CurrencyType
	ChainID               int
	ExtendedTLV           tlv.TLVList
	Address               string //chain specific attribute not stored on card
	AddressType           uint8  //chain specific address type identifier
}

func (p *Phonon) String() string {
	return fmt.Sprintf("KeyIndex: %v\nDenomination: %v\nCurrencyType: %v\nPubKey: %v\nAddress: %v\nChainID: %v\nCurveType: %v\nSchemaVersion: %v\nExtendedSchemaVersion: %v\nExtendedTLV: %v\n",
		p.KeyIndex,
		p.Denomination,
		p.CurrencyType,
		p.PubKey,
		p.Address,
		p.ChainID,
		p.CurveType,
		p.SchemaVersion,
		p.ExtendedSchemaVersion,
		p.ExtendedTLV)
}

//Phonon data structured for display to the user and use in frontends
type PhononJSON struct {
	KeyIndex              PhononKeyIndex
	PubKey                string //pubkey as hexstring
	Address               string //Chain specific address as hexstring
	AddressType           uint8
	SchemaVersion         uint8
	ExtendedSchemaVersion uint8
	Denomination          Denomination
	CurrencyType          int
	ChainID               int
	CurveType             uint8
}

//Unmarshals a PhononUserView into an internal phonon representation
func (p *Phonon) UnmarshalJSON(b []byte) error {
	phJSON := PhononJSON{}
	err := json.Unmarshal(b, &phJSON)
	if err != nil {
		return err
	}
	p.KeyIndex = phJSON.KeyIndex
	p.CurveType = CurveType(phJSON.CurveType)

	//Convert hexstring pubkey to *ecdsa.PublicKey
	pubKeyBytes, err := hex.DecodeString(phJSON.PubKey)
	if err != nil {
		return err
	}
	p.PubKey, err = NewPhononPubKey(pubKeyBytes, p.CurveType)
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
		PubKey:                p.PubKey.String(),
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
type PhononKeyIndex uint16

func KeyIndexFromBytes(keyIndexBytes []byte) PhononKeyIndex {
	return PhononKeyIndex(binary.BigEndian.Uint16(keyIndexBytes))
}
func (i PhononKeyIndex) ToBytes() []byte {
	b := make([]byte, 0)
	binary.BigEndian.PutUint16(b, uint16(i))
	return b
}

const (
	Unspecified CurrencyType = 0x0000
	Bitcoin     CurrencyType = 0x0001
	Ethereum    CurrencyType = 0x0002
	Native      CurrencyType = 0x0003
)

type CurveType uint8

const (
	Secp256k1 CurveType = iota
	NativeCurve
	Unknown = 0xFF
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
		if m.Cmp(big.NewInt(0)) == 0 {
			exponent += 1
			i.Div(i, big.NewInt(10))
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

type PhononPubKey interface {
	Decode([]byte) (PhononPubKey, error)
	String() string
	Bytes() []byte
	Equal(PhononPubKey) bool
}

type ECCPubKey struct {
	PubKey *ecdsa.PublicKey
}

func (pubKey *ECCPubKey) Decode(data []byte) (pk PhononPubKey, err error) {
	pubKey.PubKey, err = util.ParseECCPubKey(data)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (pubKey *ECCPubKey) String() string {
	return util.ECCPubKeyToHexString(pubKey.PubKey)
}

func (pubKey *ECCPubKey) Bytes() []byte {
	return crypto.FromECDSAPub(pubKey.PubKey)
}

func (pubKey *ECCPubKey) Equal(x PhononPubKey) bool {
	xx, ok := x.(*ECCPubKey)
	if !ok {
		return false
	}
	return pubKey.PubKey.Equal(xx.PubKey)
}

//Convenience function to easily convert PhononPublicKeys to their underyling concrete type when possible
func PhononPubKeyToECDSA(pubKey PhononPubKey) (*ecdsa.PublicKey, error) {
	ecc, ok := pubKey.(*ECCPubKey)
	if !ok {
		return nil, errors.New("cannot convert non-ECC pubkey to ECC")
	}
	return ecc.PubKey, nil
}

type NativePubKey struct {
	Hash []byte
}

func (nat *NativePubKey) Decode(data []byte) (pk PhononPubKey, err error) {
	if len(data) != 64 {
		log.Error("native phonon pubkey data should have been 64 bytes but was", len(data))
		return nil, errors.New("native phonon pubkey was invalid length != 64")
	}
	nat.Hash = data
	return nat, nil
}

func (nat *NativePubKey) String() string {
	return hex.EncodeToString(nat.Hash)
}

func (nat *NativePubKey) Bytes() []byte {
	return nat.Hash
}

func (nat *NativePubKey) Equal(x PhononPubKey) bool {
	xx, ok := x.(*NativePubKey)
	if !ok {
		return false
	}
	return bytes.Equal(nat.Hash, xx.Hash)
}

//NewPhononPubKey parses raw public key data into the assigned PhononPubKey interface based on the given CurveType
func NewPhononPubKey(rawPubKey []byte, crv CurveType) (pubKey PhononPubKey, err error) {
	//Switch pubkey interface based on given curveType
	switch crv {
	case Secp256k1:
		pubKey = &ECCPubKey{}
		return pubKey.Decode(rawPubKey)
	case NativeCurve:
		pubKey = &NativePubKey{}
		return pubKey.Decode(rawPubKey)
	default:
		return nil, errors.New("unknown phonon public key curve type")
	}
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
