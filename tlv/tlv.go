package tlv

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type TLV struct {
	Tag    byte
	Length int
	Value  []byte
}

type TLVCollection map[byte][][]byte

type TLVList []TLV

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
		Tag:    tag,
		Length: len(value),
		Value:  value,
	}, nil
}

//Encode a TLV structure as serialized bytes
func (tlv *TLV) Encode() []byte {
	prefix := []byte{tlv.Tag, byte(tlv.Length)}
	serializedBytes := append(prefix, tlv.Value...)
	return serializedBytes
}

/*Parses a TLV encoded response structure
Returning a flattened map where the keys are tags
and the value is a slice of raw bytes, one entry for each tag instance found.
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
		for tag, entries := range coll {
			result[tag] = append(result[tag], entries...)
		}
	}
	return result
}

//FindTag takes a tag as input and returns the first instance of the tag's value
func (coll TLVCollection) FindTag(tag byte) (value []byte, err error) {
	valueSlice, err := coll.FindTags(tag)
	if err != nil {
		return nil, err
	}
	return valueSlice[0], nil
}

//Findtags takes a tag as input and returns all instances of the tag's values as a slice of slice of byte
func (coll TLVCollection) FindTags(tag byte) (value [][]byte, err error) {
	valueSlice, exists := coll[tag]
	if !exists {
		return nil, ErrTagNotFound
	}
	if len(valueSlice) < 1 {
		return nil, ErrTagEmpty
	}
	return valueSlice, nil
}

//Takes a list of TLVs and encodes them as serialized bytes in FIFO order
func EncodeTLVList(tlvList ...TLV) []byte {
	var data []byte
	for _, tlv := range tlvList {
		data = append(data, tlv.Encode()...)
	}
	return data
}

//Removes tags from a collection, returning the remaining TLV's
//Useful for slicing out extended TLVs after the known ones are parsed
func (coll TLVCollection) GetRemainingTLVs(tags []byte) (remaining []TLV) {
	for _, tag := range tags {
		delete(coll, tag)
	}
	for tag, entry := range coll {
		/*Take the first entry
		Does not handle duplicates, since individual entries stored in
		a single phonon should only contain unique values
		Error should never trigger as we are reassembling parsed values*/
		remainingTLV, _ := NewTLV(tag, entry[0])

		remaining = append(remaining, remainingTLV)
	}
	return
}

func (tlv TLV) String() string {
	return fmt.Sprintf("Tag: % X, Length: %v, Value: % X", tlv.Tag, tlv.Length, tlv.Value)
}

func (tlvlist TLVList) String() string {
	ret := ""
	for _, tlv := range tlvlist {
		ret = ret + " " + tlv.String()
	}
	return ret
}
