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
var ErrTagNotFound = errors.New("tag not found in TLV collection")
var ErrTagEmpty = errors.New("tag contained no parsed data")

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

/*Parses a TLV encoded response structure
Returning a flattened map where the keys are tags
and the value is a slice of raw bytes, one entry for each tag instance found
For any "constructedTags" passed, the parser will recurse into the value of that
tag to find internal TLV's and append them to the collection as flattened entries */
func ParseTLVPacket(data []byte, constructedTags ...byte) (TLVCollection, error) {
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
		for _, constructedTag := range constructedTags {
			if tag == constructedTag {
				nestedResult, err := ParseTLVPacket(value, constructedTags...)
				if err != nil {
					return result, err
				}
				result = mergeTLVCollections(result, nestedResult)
			}
		}
	}

}

func mergeTLVCollections(collections ...TLVCollection) TLVCollection {
	result := TLVCollection{}
	for _, coll := range collections {
		for tag, value := range coll {
			for _, entry := range value {
				result[tag] = append(result[tag], entry)
			}
		}
	}
	return result
}

//FindTag takes a tag as input and returns the first instance of the tag
func (coll TLVCollection) FindTag(tag byte) (value []byte, err error) {
	valueSlice, exists := coll[tag]
	if !exists {
		return nil, ErrTagNotFound
	}

	if len(valueSlice) < 1 {
		return nil, ErrTagEmpty
	}
	return valueSlice[0], nil
}

var SetDescriptorResponse struct {
	PhononKeyIndex  TLV
	PhononPublicKey TLV
}
