package util

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"math/big"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type ECDSASignature struct {
	R, S *big.Int
}

func ParseECDSASignature(rawSig []byte) (*ECDSASignature, error) {
	signature := &ECDSASignature{}
	_, err := asn1.Unmarshal(rawSig, signature)
	if err != nil {
		log.Error("could not unmarshal raw signature into ECDSA format: ", err)
		log.Error("raw sig:\n", hex.Dump(rawSig))
		return nil, err
	}
	return signature, nil
}

func ParseECDSAPubKey(pubKey []byte) (*ecdsa.PublicKey, error) {
	ecdsaPubKey, err := ethcrypto.UnmarshalPubkey(pubKey)
	if err != nil {
		log.Error("could not unmarshal ecdsa pub key from raw: ", err)
		log.Error("raw pubkey:\n", hex.Dump(pubKey))
		return nil, err
	}
	return ecdsaPubKey, nil
}

func SerializeECDSAPubKey(pubKey *ecdsa.PublicKey) []byte {
	var ECC_POINT_FORMAT_UNCOMPRESSED byte = 0x04
	pubKeyBytes := []byte{ECC_POINT_FORMAT_UNCOMPRESSED}
	pubKeyBytes = append(pubKeyBytes, pubKey.X.Bytes()...)
	pubKeyBytes = append(pubKeyBytes, pubKey.Y.Bytes()...)

	return pubKeyBytes
}

func ParseECCPrivKey(privKey []byte) (*ecdsa.PrivateKey, error) {
	eccPrivKey, err := ethcrypto.ToECDSA(privKey)
	if err != nil {
		log.Error("could not parse ecc priv key from raw bytes: ", err)
		return nil, err
	}
	return eccPrivKey, nil
}
