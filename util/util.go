package util

import "crypto/rand"

func RandomKey(length int) []byte {
	key := make([]byte, length)
	rand.Read(key)
	return key
}
