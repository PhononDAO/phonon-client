package card

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/GridPlus/keycard-go/crypto"

	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/util"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

//Insecure alternative implementation of phonon command set which uses static keys in the card to terminal pairing process
//in order to allow for debugging with the javacard simulator
type StaticPhononCommandSet struct {
	*PhononCommandSet
}

func NewStaticPhononCommandSet(cs *PhononCommandSet) *StaticPhononCommandSet {
	return &StaticPhononCommandSet{
		cs,
	}
}

func (cs *StaticPhononCommandSet) Select() (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error) {
	cmd := NewCommandSelectPhononApplet()
	cmd.ApduCmd.SetLe(0)

	log.Debug("sending static SELECT command")
	resp, err := cs.Send(cmd)
	if err != nil {
		log.Error("could not send select command. err: ", err)
		return nil, nil, false, err
	}
	err = cs.checkOK(resp, err)
	if err != nil {
		return nil, nil, false, err
	}
	instanceUID, cardPubKey, cardInitialized, err = parseSelectResponse(resp.Data)
	if err != nil {
		log.Error("error parsing select response. err: ", err)
		return nil, nil, false, err
	}

	//Generate secure channel secrets using card's public key
	secretsErr := cs.sc.GenerateStaticSecret(ethcrypto.FromECDSAPub(cardPubKey))
	if secretsErr != nil {
		log.Error("could not generate secure channel secrets. err: ", secretsErr)
		return nil, nil, true, secretsErr
	}
	log.Debugf("Pairing generated key: % X\n", cs.sc.RawPublicKey())

	return instanceUID, cardPubKey, cardInitialized, nil
}

func (cs *StaticPhononCommandSet) Pair() (*cert.CardCertificate, error) {
	log.Debug("sending static PAIR command")
	//Generate static salt
	clientSalt := make([]byte, 0)
	for i := 0; i < 32; i++ {
		clientSalt = append(clientSalt, 0x01)
	}

	//Staticize this key
	staticEntropy := make([]byte, 256)
	for i := 0; i < 256; i++ {
		staticEntropy = append(staticEntropy, 0x01)
	}
	r := strings.NewReader(string(staticEntropy))
	pairingPrivKey, err := ecdsa.GenerateKey(ethcrypto.S256(), r)
	if err != nil {
		log.Error("unable to generate pairing keypair. err: ", err)
		return &cert.CardCertificate{}, err
	}
	pairingPubKey := pairingPrivKey.PublicKey

	//Exchange pairing key info with card
	cmd := NewCommandPairStep1(clientSalt, &pairingPubKey)

	resp, err := cs.Send(cmd)
	if err != nil {
		log.Error("unable to send Pair Step 1 command. err: ", err)
		return &cert.CardCertificate{}, err
	}
	err = checkPairingErrors(1, resp.Sw)
	if err != nil {
		return &cert.CardCertificate{}, err
	}

	salt, cardCert, signature, err := ParsePairStep1Response(resp.Data)
	if err != nil {
		log.Error("could not parse pair step 1 response. err: ", err)
		return &cert.CardCertificate{}, err
	}

	cardCertPubKey, err := util.ParseECDSAPubKey(cardCert.PubKey)
	if err != nil {
		return &cert.CardCertificate{}, err
	}
	//Validate card's certificate has valid GridPlus signature
	certValid := cert.ValidateCardCertificate(cardCert, gridplus.SafecardDevCAPubKey)
	log.Debug("certificate signature valid: ", certValid)
	if !certValid {
		log.Error("unable to verify card certificate.")
		return &cert.CardCertificate{}, err
	}

	pubKeyValid := gridplus.ValidateECCPubKey(cardCertPubKey)
	log.Debug("certificate public key valid: ", pubKeyValid)
	if !pubKeyValid {
		log.Error("card pubkey invalid")
		return &cert.CardCertificate{}, err
	}

	//challenge message test
	ecdhSecret := crypto.GenerateECDHSharedSecret(pairingPrivKey, cardCertPubKey)

	secretHashArray := sha256.Sum256(append(clientSalt, ecdhSecret...))
	secretHash := secretHashArray[0:]

	//validate that card created valid signature over same salted and hashed ecdh secret
	valid := ecdsa.VerifyASN1(cardCertPubKey, secretHash, signature)
	if !valid {
		log.Error("ecdsa sig not valid")
		return &cert.CardCertificate{}, errors.New("could not verify shared secret challenge")
	}
	cryptogram := sha256.Sum256(append(salt, secretHash...))

	log.Debug("sending PAIR step 2 cmd")
	cmd = NewCommandPairStep2(cryptogram)
	resp, err = cs.Send(cmd)
	if err != nil {
		log.Error("error sending pair step 2 command. err: ", err)
		return &cert.CardCertificate{}, err
	}

	err = checkPairingErrors(2, resp.Sw)
	if err != nil {
		return &cert.CardCertificate{}, err
	}
	pairStep2Resp, err := gridplus.ParsePairStep2Response(resp.Data)
	if err != nil {
		log.Error("could not parse pair step 2 response. err: ", err)
		return &cert.CardCertificate{}, err
	}
	log.Debugf("pairStep2Resp: % X", pairStep2Resp)

	//Derive Pairing Key
	pairingKey := sha256.Sum256(append(pairStep2Resp.Salt, secretHash...))
	log.Debugf("derived pairing key: % X", pairingKey)

	//Store pairing info for use in OpenSecureChannel
	cs.setPairingInfo(pairingKey[0:], pairStep2Resp.PairingIdx)

	log.Debug("pairing succeeded")
	return &cardCert, nil
}
