package card

import (
	"bytes"
	"errors"
	"io"
)

type TLV struct {
	tag    byte
	length int
	value  []byte
}

type TLVCollection map[byte][][]byte

const MaxValueBytes = 256

var ErrValueLengthExceedsMax = errors.New("value exceeds max allowable length")
var ErrDataNotFound = errors.New("data read hit EOF before specified length was reached")

//Create a TLV struct from a tag identifier and a value represented as bytes
func NewTLV(tag byte, value []byte) (TLV, error) {
	if len(value) > MaxValueBytes {
		return TLV{}, ErrValueLengthExceedsMax
	}
	return TLV{
		tag:    tag,
		length: len(value),
		value:  value,
	}, nil
}

//Encode a TLV structure as serialized bytes
func (tlv *TLV) Encode() []byte {
	prefix := []byte{tlv.tag, byte(tlv.length)}
	serializedBytes := append(prefix, tlv.value...)
	return serializedBytes
}

func ParseTLVPacket(data []byte) (TLVCollection, error) {
	buf := bytes.NewBuffer(data)
	result := make(TLVCollection)

	for {
		tag, err := buf.ReadByte()
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return result, err
		}
		length, err := buf.ReadByte()
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return result, err
		}
		value := make([]byte, int(length))
		_, err = buf.Read(value)
		if err != nil {
			return result, ErrDataNotFound
		}
		result[tag] = append(result[tag], value)
	}

}

var SetDescriptorResponse struct {
	PhononKeyIndex  TLV
	PhononPublicKey TLV
}
