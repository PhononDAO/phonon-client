package cert

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
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	log "github.com/sirupsen/logrus"
)

type CardCertificate struct {
	Permissions CertPermissions
	PubKey      []byte
	Sig         []byte
}

type CertPermissions struct {
	certType    byte
	certLen     byte
	permType    byte
	permLen     byte
	permissions []byte
	pubKeyType  byte
	pubKeyLen   byte
}

// Dev cert CA Key
var PhononDemoCAPubKey = []byte{
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

var PhononAlphaCAPubKey = []byte{
	0x04,
	0x72, 0xd5, 0x8c, 0x1e, 0xc4, 0x8f, 0x00, 0x72,
	0xeb, 0xc3, 0x97, 0x12, 0xa8, 0xec, 0x74, 0xe5,
	0xa4, 0x58, 0x19, 0x31, 0xd6, 0xff, 0xe5, 0x97,
	0xb6, 0x45, 0x9b, 0x46, 0x3c, 0x87, 0xfc, 0xe1,
	0x59, 0xb8, 0xe1, 0xae, 0x40, 0xc3, 0x83, 0xcd,
	0xae, 0x78, 0xaa, 0xdf, 0xff, 0xb0, 0x83, 0x91,
	0x7c, 0x91, 0x1c, 0x3f, 0x9d, 0x75, 0xa5, 0xf1,
	0xa9, 0x24, 0xb6, 0x27, 0xf1, 0x5d, 0xec, 0x51,
}

//Additional CA Key for testing purposes
var PhononMockCAPubKey = []byte{
	0x04,
	0xa0, 0x48, 0xd2, 0x7a, 0xe0, 0x10, 0xeb, 0x05,
	0x82, 0x32, 0x25, 0xd9, 0x8a, 0x00, 0xf8, 0x19,
	0xe7, 0x93, 0x88, 0x08, 0xf4, 0x04, 0x40, 0x0b,
	0x4a, 0x8b, 0x66, 0xc3, 0x09, 0xa7, 0x54, 0x15,
	0x80, 0x81, 0xc8, 0x09, 0x3b, 0x49, 0x19, 0xe4,
	0x13, 0x69, 0x48, 0x33, 0xc1, 0x60, 0xe7, 0xcf,
	0x3b, 0x77, 0x92, 0xd6, 0x73, 0x8c, 0xce, 0x54,
	0x6b, 0xf0, 0x67, 0x99, 0x7b, 0x18, 0x0f, 0x11,
}

var PhononMockCAPrivKey = []byte{
	0xab, 0x7e, 0xa6, 0xe2, 0xa6, 0xcf, 0x1c, 0x7f,
	0xb4, 0xb8, 0x5b, 0x43, 0xba, 0x47, 0x2a, 0x85,
	0xfd, 0x94, 0xd6, 0x9b, 0x67, 0xfa, 0xce, 0x7a,
	0x9a, 0x07, 0xcd, 0xde, 0x16, 0x85, 0xd8, 0x3b,
}

var ErrInvalidCert = errors.New("certificate signature was invalid")

//Accepts a safecard certificate and validates it against the provided CA PubKey
//Safecard CA's provided by SafecardProdCAPubKey or SafecardDevCAPubKey for the respective environments
func ValidateCardCertificate(cert CardCertificate, CAPubKey []byte) error {
	//Hash of cert excepting signature, certType, and certLen
	certBytes := cert.Digest()
	certHash := sha256.Sum256(certBytes)

	CApubKey, err := util.ParseECCPubKey(CAPubKey)
	if err != nil {
		log.Error("could not parse CAPubKey: ", err)
		return err
	}
	log.Debug("certificate CA PubKey was valid")
	signature, err := util.ParseECDSASignature(cert.Sig)
	if err != nil {
		log.Error("could not parse cert signature: ", err)
		return err
	}

	valid := ecdsa.Verify(CApubKey, certHash[0:], signature.R, signature.S)
	if !valid {
		return ErrInvalidCert
	}
	return nil
}

//Create a card certificate, signing with the key supplied in the signKeyFunc
func CreateCardCertificate(cardPubKey *ecdsa.PublicKey, signKeyFunc func([]byte) ([]byte, error)) ([]byte, error) {
	cardPubKeyBytes := ethcrypto.FromECDSAPub(cardPubKey)

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
	key.Curve = secp256k1.S256()
	key.X, key.Y = key.ScalarBaseMult(key.D.Bytes())
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

		c := connector.NewHTTPConnector("localhost:12345")
		sm, err := yubihsm.NewSessionManager(c, 1, password)
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

		return res.(*commands.SignDataEcdsaResponse).Signature, nil
	}
}

func GetSignerWithPrivateKey(privKey ecdsa.PrivateKey) func([]byte) ([]byte, error) {
	return func(cert []byte) ([]byte, error) {
		digest := sha256.Sum256(cert)
		ret, err := privKey.Sign(rand.Reader, digest[:], nil)
		if err != nil {
			return []byte{}, err
		}
		log.Debug("finished signing")
		return ret, nil
	}
}

func ParseRawCardCertificate(cardCertificateRaw []byte) (cert CardCertificate, err error) {
	if len(cardCertificateRaw) < 4 {
		return CardCertificate{}, errors.New("card certificate length too short to read permissions length")
	}
	cert.Permissions.certType = cardCertificateRaw[0]
	cert.Permissions.certLen = cardCertificateRaw[1]
	cert.Permissions.permType = cardCertificateRaw[2]
	cert.Permissions.permLen = cardCertificateRaw[3]

	if cert.Permissions.certLen == 0 || cert.Permissions.permLen == 0 {
		log.Debugf("invalid certificate found: % X", cardCertificateRaw)
		return CardCertificate{}, errors.New("card certificate invalid")
	}
	permsLen := int(cert.Permissions.permLen)
	if len(cardCertificateRaw) < 5+permsLen {
		return CardCertificate{}, errors.New("card certificate too short to read full permissions block")
	}
	cert.Permissions.permissions = cardCertificateRaw[4 : 4+permsLen]
	cert.Permissions.pubKeyType = cardCertificateRaw[4+permsLen]
	cert.Permissions.pubKeyLen = cardCertificateRaw[5+permsLen]
	pubKeyLen := int(cert.Permissions.pubKeyLen)
	certLength := int(cert.Permissions.certLen)
	if len(cardCertificateRaw) < certLength {
		return CardCertificate{}, errors.New("card certificate was incorrect length")
	}
	if len(cardCertificateRaw) < 6+permsLen+pubKeyLen {
		return CardCertificate{}, errors.New("card certificate incorrect length")
	}
	cert.PubKey = cardCertificateRaw[6+int(cert.Permissions.permLen) : 6+permsLen+pubKeyLen]
	cert.Sig = cardCertificateRaw[6+permsLen+pubKeyLen : certLength]

	return cert, nil
}

//Digest the certificate data, permissions and pubkey into bytes
//This is the set of bytes used to sign and validate the certificate
//(skips the first two bytes for cert type and length)
func (cert CardCertificate) Digest() []byte {
	bytes := []byte{
		cert.Permissions.permType,
		cert.Permissions.permLen,
	}
	bytes = append(bytes, cert.Permissions.permissions...)
	bytes = append(bytes, cert.Permissions.pubKeyType)
	bytes = append(bytes, cert.Permissions.pubKeyLen)
	bytes = append(bytes, cert.PubKey...)
	return bytes
}

//Serialize the full certificate, including the cert type and length
//which are unused in the certificate signature
func (cert CardCertificate) Serialize() []byte {
	bytes := []byte{
		cert.Permissions.certType,
		cert.Permissions.certLen,
	}
	bytes = append(bytes, cert.Digest()...)
	bytes = append(bytes, cert.Sig...)
	return bytes
}

func (cert CardCertificate) String() string {
	return fmt.Sprintf("Permissions: % X, PubKey: % X (length: %v), Sig: % X (length %v)", cert.Permissions, cert.PubKey, len(cert.PubKey), cert.Sig, len(cert.Sig))
}
