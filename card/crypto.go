package card

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v2"
	log "github.com/sirupsen/logrus"

	"github.com/GridPlus/phonon-client/util"
)

func validateCardCertificate(cert SafecardCert) bool {
	//TODO: break up the parsing of these values into their own
	//functions which make some sense so I can reuse this later

	//Hash of cert bytes,
	//this is correct
	certBytes := append(cert.permissions, cert.pubKey...)
	certHash := sha256.Sum256(certBytes)

	//Components of CA certificate public key
	X := new(big.Int)
	Y := new(big.Int)
	X.SetBytes(SafecardCertCAPubKey[1:33])
	Y.SetBytes(SafecardCertCAPubKey[33:])

	CApubKey := &ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     X,
		Y:     Y,
	}

	log.Infof("signature: length: %v\n% X", len(cert.sig), cert.sig)
	//raw sig field is correct

	//Able to decode the DER signature with this library, should do more of this.
	type ECDSASignature struct {
		R, S *big.Int
	}
	signature := &ECDSASignature{}
	_, err := asn1.Unmarshal(cert.sig, signature)
	if err != nil {
		log.Error("could not unmarshal certificate signature.", err)
	}

	log.Infof("certHash: % X", certHash)
	log.Info("pubKey X ", X)
	log.Info("pubKey Y ", Y)

	return ecdsa.Verify(CApubKey, certHash[0:], signature.R, signature.S)
}

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

func computeECDHSharedSecret(clientSalt []byte, privKey *secp256k1.PrivateKey, safecardSalt []byte, cardPubKey *ecdsa.PublicKey, cardSig []byte) (secretHash []byte, cryptogram []byte, err error) {
	//compute a shared secret
	//Using the secp256k1 library implementation
	secpCardPubKey := secp256k1.NewPublicKey(cardPubKey.X, cardPubKey.Y)
	// log.Infof("secpCardPubKey: % X", secpCardPubKey)

	ecdhSecret := secp256k1.GenerateSharedSecret(privKey, secpCardPubKey)
	if err != nil {
		log.Error("could not compute shared secret. err: ", err)
		return nil, nil, err
	}
	// log.Infof("ecdhSecret: % X", ecdhSecret)

	log.Info("raw cardSig length: ", len(cardSig))
	log.Infof("raw cardSig: % X", cardSig)
	// secretHash = sha256(clientSalt, ECDH secret)
	secretHashArray := sha256.Sum256(append(clientSalt, ecdhSecret...))
	secpCardSig, err := secp256k1.ParseDERSignature(cardSig)
	if err != nil {
		log.Error("invalid card sig on shared secret. err: ", err)
		return nil, nil, err
	}
	// log.Info("orignal safecardCert.sig length: ", len(cardSig))
	// log.Infof("original safecardCert.sig payload:\n% X", cardSig)

	sharedSecretValid := secpCardSig.Verify(secretHashArray[0:], secpCardPubKey)
	if !sharedSecretValid {
		log.Error("could not verify card signature on challenge message")
		return nil, nil, err
	}
	log.Info("shared secret challenge valid: ", sharedSecretValid)

	//Compute Client Crytogram
	cryptogramArray := sha256.Sum256(append(safecardSalt, secretHashArray[0:]...))
	return secretHashArray[0:], cryptogramArray[0:], nil
}

func validateECCPubKey(pubKey *ecdsa.PublicKey) bool {
	if !pubKey.IsOnCurve(pubKey.X, pubKey.Y) {
		log.Error("pubkey is not valid point on curve")
		return false
	}

	//TODO: more checks for point is not at infinity, not sure if these are needed
	return true
}
