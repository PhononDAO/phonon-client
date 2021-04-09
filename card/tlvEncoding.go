package card

import "encoding/binary"

func EncodeKeyIndexList(keyIndices []uint16) []byte {
	var keyIndexBytes []byte
	b := make([]byte, 2)
	for _, keyIndex := range keyIndices {
		binary.BigEndian.PutUint16(b, keyIndex)
		keyIndexBytes = append(keyIndexBytes, b...)
	}
	//TODO: possibly handle potential error
	data, _ := NewTLV(TagPhononKeyIndexList, keyIndexBytes)
	return data.Encode()
}
