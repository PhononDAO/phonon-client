package util

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
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

func ECDSAPubKeyToHexString(pubKey *ecdsa.PublicKey) string {
	return fmt.Sprintf("%x", ethcrypto.FromECDSAPub(pubKey))
}

func ParseECCPrivKey(privKey []byte) (*ecdsa.PrivateKey, error) {
	eccPrivKey, err := ethcrypto.ToECDSA(privKey)
	if err != nil {
		log.Error("could not parse ecc priv key from raw bytes: ", err)
		return nil, err
	}
	return eccPrivKey, nil
}
