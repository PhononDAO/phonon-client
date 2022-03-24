package util

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type ECDSASignature struct {
	R, S *big.Int
}

var ErrInvalidECCPubKeyFormat = errors.New("ECC pubkey format could not be detected")

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

func ParseECCPubKey(rawPubKey []byte) (pubKey *ecdsa.PublicKey, err error) {
	if len(rawPubKey) == 0 {
		return nil, errors.New("pubKey was zero length")
	}
	if rawPubKey[0] == 0x04 {
		//Unmarshal uncompressed pubkey format
		pubKey, err = ethcrypto.UnmarshalPubkey(rawPubKey)
		if err != nil {
			log.Error("could not unmarshal uncompressed ecdsa pub key from raw: ", err)
			log.Error("raw pubkey:\n", hex.Dump(rawPubKey))
			return nil, err
		}
	} else if rawPubKey[0] == 0x02 || rawPubKey[0] == 0x03 {
		//Unmarshal compressed pubkey format
		pubKey, err = ethcrypto.DecompressPubkey(rawPubKey)
		if err != nil {
			log.Error("could not unmarshal compressed ecdsa pub key from raw: ", err)
			return nil, err
		}
	} else {
		log.Debugf("could not detect ECC pubkey format from key: % X\n", rawPubKey)
		return nil, ErrInvalidECCPubKeyFormat
	}

	return pubKey, nil
}

func ECCPubKeyToHexString(pubKey *ecdsa.PublicKey) string {
	return fmt.Sprintf("%x", ethcrypto.FromECDSAPub(pubKey))
}

func ECCPrivKeyToHex(privKey *ecdsa.PrivateKey) string {
	return fmt.Sprintf("%x", ethcrypto.FromECDSA(privKey))
}

func ParseECCPrivKey(privKey []byte) (*ecdsa.PrivateKey, error) {
	eccPrivKey, err := ethcrypto.ToECDSA(privKey)
	if err != nil {
		log.Error("could not parse ecc priv key from raw bytes: ", err)
		return nil, err
	}
	return eccPrivKey, nil
}

func CardIDFromPubKey(pubKey *ecdsa.PublicKey) string {
	return ECCPubKeyToHexString(pubKey)[:16]
}
