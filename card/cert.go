package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/GridPlus/phonon-client/util"
	yubihsm "github.com/certusone/yubihsm-go"
	"github.com/certusone/yubihsm-go/commands"
	"github.com/certusone/yubihsm-go/connector"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	log "github.com/sirupsen/logrus"
)

type CardCertificate struct {
	Permissions []byte
	PubKey      []byte
	Sig         []byte
}

// Dev cert CA Key
var SafecardDevCAPubKey = []byte{
	0x04,
	0x5c, 0xfd, 0xf7, 0x7a, 0x00, 0xb4, 0xb6, 0xb4,
	0xa5, 0xb8, 0xbb, 0x26, 0xb5, 0x49, 0x7d, 0xbc,
	0x7a, 0x4d, 0x01, 0xcb, 0xef, 0xd7, 0xaa, 0xea,
	0xf5, 0xf6, 0xf8, 0xf8, 0x86, 0x59, 0x76, 0xe7,
	0x94, 0x1a, 0xb0, 0xec, 0x16, 0x51, 0x20, 0x9c,
	0x44, 0x40, 0x09, 0xfd, 0x48, 0xd9, 0x25, 0xa1,
	0x7d, 0xe5, 0x04, 0x0b, 0xa4, 0x7e, 0xaf, 0x3f,
	0x5b, 0x51, 0x72, 0x0d, 0xd4, 0x0b, 0x2f, 0x9d,
}

// Prod cert CA Key
var SafecardProdCAPubKey = []byte{
	0x04,
	0x77, 0x81, 0x6e, 0x8e, 0x83, 0xbb, 0x17, 0xc4,
	0x30, 0x9c, 0xc2, 0xe5, 0xaa, 0x13, 0x4c, 0x57,
	0x3a, 0x59, 0x43, 0x15, 0x49, 0x40, 0x09, 0x5a,
	0x42, 0x31, 0x49, 0xf7, 0xcc, 0x03, 0x84, 0xad,
	0x52, 0xd3, 0x3f, 0x1b, 0x4c, 0xd8, 0x9c, 0x96,
	0x7b, 0xf2, 0x11, 0xc0, 0x39, 0x20, 0x2d, 0xf3,
	0xa7, 0x89, 0x9c, 0xb7, 0x54, 0x3d, 0xe4, 0x73,
	0x8c, 0x96, 0xa8, 0x1c, 0xfd, 0xe4, 0xb1, 0x17,
}

//Accepts a safecard certificate and validates it against the provided CA PubKey
//Safecard CA's provided by SafecardProdCAPubKey or SafecardDevCAPubKey for the respective environments
func ValidateCardCertificate(cert CardCertificate, CAPubKey []byte) bool {
	//Hash of cert excepting signature
	certBytes := append(cert.Permissions, cert.PubKey...)
	certHash := sha256.Sum256(certBytes)

	CApubKey, err := util.ParseECDSAPubKey(CAPubKey)
	if err != nil {
		log.Error("could not parse CAPubKey: ", err)
		return false
	}
	signature, err := util.ParseECDSASignature(cert.Sig)
	if err != nil {
		log.Error("could not parse cert signature: ", err)
		return false
	}

	log.Debugf("certHash: % X", certHash)

	return ecdsa.Verify(CApubKey, certHash[0:], signature.R, signature.S)
}

//Create a card certificate, signing with the key supplied in the signKeyFunc
func createCardCertificate(cardPubKey *ecdsa.PublicKey, signKeyFunc func([]byte) ([]byte, error)) ([]byte, error) {
	cardPubKeyBytes := util.SerializeECDSAPubKey(cardPubKey)

	// Create Card Certificate
	perms := []byte{0x30, 0x00, 0x02, 0x02, 0x00, 0x00, 0x80, 0x41}
	cardCertificate := append(perms, cardPubKeyBytes...)

	// Sign The Certificate
	preImage := cardCertificate[2:]
	sig, err := signKeyFunc(preImage)
	if err != nil {
		return nil, fmt.Errorf("unable to sign Cert: %s", err.Error())
	}

	// Append CA Signature to certificate
	signedCert := append(cardCertificate, sig...)

	//Substitute actual certificate length in certificate header
	signedCert[1] = byte(len(signedCert))
	return signedCert, nil
}

func SignWithDemoKey(cert []byte) ([]byte, error) {
	demoKey := []byte{
		0x03, 0x8D, 0x01, 0x08, 0x90, 0x00, 0x00, 0x00,
		0x10, 0xAA, 0x82, 0x07, 0x09, 0x80, 0x00, 0x00,
		0x01, 0xBB, 0x03, 0x06, 0x90, 0x08, 0x35, 0xF9,
		0x10, 0xCC, 0x04, 0x85, 0x09, 0x00, 0x00, 0x91,
	}
	var key ecdsa.PrivateKey

	// print("signing with demo key")
	key.D = new(big.Int).SetBytes(demoKey)
	key.PublicKey.Curve = secp256k1.S256()
	key.PublicKey.X, key.PublicKey.Y = key.PublicKey.Curve.ScalarBaseMult(key.D.Bytes())
	digest := sha256.Sum256(cert)
	ret, err := key.Sign(rand.Reader, digest[:], nil)
	if err != nil {
		return []byte{}, err
	}
	println("finished signing")
	return ret, nil
}

func SignWithYubikeyFunc(slot int, password string) func([]byte) ([]byte, error) {
	return func(cert []byte) ([]byte, error) {

		c := connector.NewHTTPConnector("localhost:1234")
		sm, err := yubihsm.NewSessionManager(c, 1, "password")
		if err != nil {
			panic(err)
		}

		digest := sha256.Sum256(cert)
		uint16slot := uint16(slot)
		cmd, err := commands.CreateSignDataEcdsaCommand(uint16slot, digest[:])
		if err != nil {
			return []byte{}, err
		}
		res, err := sm.SendEncryptedCommand(cmd)
		if err != nil {
			return []byte{}, err
		}

		return res.(*commands.SignDataEddsaResponse).Signature, nil
	}
}

func ParseRawCertificate(cardCertificateRaw []byte) (CardCertificate, error) {
	certLength := int(cardCertificateRaw[1])
	if len(cardCertificateRaw) < certLength {
		return CardCertificate{}, errors.New("certificate was incorrect length")
	}
	cardCertificate := CardCertificate{
		Permissions: cardCertificateRaw[2:8],
		PubKey:      cardCertificateRaw[8 : 8+65],
		Sig:         cardCertificateRaw[8+65 : 0+certLength],
	}
	return cardCertificate, nil
}
