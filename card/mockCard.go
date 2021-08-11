package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"unicode"

	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type MockCard struct {
	Phonons        []model.Phonon
	pin            string
	pinVerified    bool
	sc             SecureChannel
	identityKey    *ecdsa.PrivateKey
	IdentityPubKey *ecdsa.PublicKey
	IdentityCert   []byte
}

func NewMockCard() (*MockCard, error) {
	identityPrivKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &MockCard{
		identityKey:    identityPrivKey,
		IdentityPubKey: &identityPrivKey.PublicKey,
	}, nil
}

func (c *MockCard) Select() (instanceUID []byte, cardPubKey []byte, cardInitialized bool, err error) {
	instanceUID = util.RandomKey(16)

	privKey, _ := ethcrypto.GenerateKey()
	cardPubKey = ethcrypto.FromECDSAPub(&privKey.PublicKey)

	if c.pin == "" {
		cardInitialized = true
	} else {
		cardInitialized = false
	}
	return instanceUID, cardPubKey, true, nil
}

//PIN functions
func validatePIN(pin string) error {
	if len(pin) != 6 {
		return errors.New("pin must be 6 digits")
	}
	for _, char := range pin {
		if !unicode.IsDigit(char) {
			return errors.New("pin contained characters not in range [0-9]")
		}
	}
	return nil
}

func (c *MockCard) Init(pin string) error {
	if c.pin != "" {
		return errors.New("pin already initialized")
	}
	if err := validatePIN(pin); err != nil {
		return err
	}
	c.pin = pin
	return nil
}

func (c *MockCard) VerifyPIN(pin string) error {
	if c.pin == "" {
		return errors.New("pin not initialized")
	}
	if pin != c.pin {
		c.pinVerified = false
		return errors.New("pin did not match")
	}
	c.pinVerified = true
	return nil
}

func (c *MockCard) ChangePIN(pin string) error {
	if !c.pinVerified {
		return errors.New("pin not verified")
	}
	err := validatePIN(pin)
	if err != nil {
		return err
	}
	c.pin = pin
	return nil
}

func (c *MockCard) IdentifyCard(nonce []byte) (cardPubKey *ecdsa.PublicKey, cardSig *util.ECDSASignature, err error) {
	rawCardSig, err := ecdsa.SignASN1(rand.Reader, c.identityKey, nonce)
	if err != nil {
		return c.IdentityPubKey, nil, err
	}
	cardSig, err = util.ParseECDSASignature(rawCardSig)
	if err != nil {
		return c.IdentityPubKey, nil, err
	}
	return c.IdentityPubKey, cardSig, nil
}

func (c *MockCard) InstallCertificate(signKeyFunc func([]byte) ([]byte, error)) error {
	var err error
	c.IdentityCert, err = createCardCertificate(c.IdentityPubKey, signKeyFunc)
	if err != nil {
		return err
	}
	return nil
}

func (c *MockCard) InitCardPairing() (initPairingData []byte, err error) {
	cardCertTLV, err := NewTLV(TagCardCertificate, c.IdentityCert)
	if err != nil {
		return nil, err
	}
	salt, err := NewTLV(TagSalt, util.RandomKey(32))
	if err != nil {
		return nil, err
	}
	initPairingData = EncodeTLVList(cardCertTLV, salt)

	return initPairingData, nil
}

func (c *MockCard) CardPair(initCardPairingData []byte) (cardPairingData []byte, err error) {
	//Initialize pairing salt
	receiverSalt := util.RandomKey(32)

	//Parse Pairing Values from counterparty
	tlv, err := ParseTLVPacket(initCardPairingData)
	if err != nil {
		return nil, errors.New("could not parse TLV packet")
	}
	senderCardCertRaw, err := tlv.FindTag(TagCardCertificate)
	if err != nil {
		return nil, errors.New("could not find certificate tlv tag")
	}
	senderSalt, err := tlv.FindTag(TagSalt)
	if err != nil {
		return nil, errors.New("could not find sender salt tlv tag")
	}

	senderCardCert, err := ParseRawCardCertificate(senderCardCertRaw)
	if err != nil {
		return nil, err
	}
	senderPubKey, err := util.ParseECDSAPubKey(senderCardCert.PubKey)
	if err != nil {
		return nil, err
	}
	log.Debug("certificate length: ", len(senderCardCertRaw))
	log.Debugf("% X", senderCardCertRaw)
	log.Debug("length of Permissions: ", len(senderCardCert.Permissions))
	log.Debugf("Permissions: % X", senderCardCert.Permissions)
	log.Debug("length of PubKey: ", len(senderCardCert.PubKey))
	log.Debugf("PubKey: % X", senderCardCert.PubKey)
	log.Debug("length of Sig: ", len(senderCardCert.Sig))
	log.Debugf("Sig: % X", senderCardCert.Sig)

	//Validate counterparty certificate
	valid := ValidateCardCertificate(senderCardCert, gridplus.SafecardDevCAPubKey)
	if !valid {
		return nil, errors.New("counterparty certificate signature was invalid")
	}

	pubKeyValid := gridplus.ValidateECCPubKey(senderPubKey)
	if !pubKeyValid {
		return nil, errors.New("counterparty public key is not valid ECC point")
	}

	//Compute shared secret
	ecdhSecret := crypto.GenerateECDHSharedSecret(c.identityKey, senderPubKey)

	//Compute session key with salts from both parties and ECDH secret
	sessionKeyMaterial := append(senderSalt, receiverSalt...)
	sessionKeyMaterial = append(sessionKeyMaterial, ecdhSecret...)

	sessionKey := sha512.Sum512(sessionKeyMaterial)

	//Derive secure channel info
	//Needed for TODO
	// encKey := sessionKey[:len(sessionKey)/2]
	// mac := sessionKey[len(sessionKey)/2:]

	aesIV := util.RandomKey(16)

	//TODO: Establish Secure Channel with encKey, mac, and aesIV

	//Combine shared derived session key with randomly generated aesIV and sign to prove possession of the
	//private key corresponding to the public key which established this channel's foundational ECDH secret
	cryptogram := sha256.Sum256(append(sessionKey[0:], aesIV...))
	receiverSig, err := ecdsa.SignASN1(rand.Reader, c.identityKey, cryptogram[0:])
	if err != nil {
		return nil, err
	}
	cardPairingData = append(c.IdentityCert, util.SerializeECDSAPubKey(c.IdentityPubKey)...)
	cardPairingData = append(cardPairingData, receiverSalt...)
	cardPairingData = append(cardPairingData, aesIV...)
	cardPairingData = append(cardPairingData, receiverSig...)
	return cardPairingData, nil
}

//Phonon Management Functions
//TODO
func (c *MockCard) CreatePhonons(n int) (pubKeys [][]byte, err error) {
	phononPubKeys := make([][]byte, 0)
	for i := 0; i < n; i++ {
		//65 bytes ECC key
		phononPubKeys = append(phononPubKeys, util.RandomKey(65))
	}
	return phononPubKeys, nil
}

//TODO
func (c *MockCard) SetDescriptors(phonons []model.Phonon) error {
	c.Phonons = append(c.Phonons, phonons...)
	return nil
}

//TODO
func (c *MockCard) OpenChannel() (string, error) {
	//not implemented
	return "", nil
}

//TODO
func (c *MockCard) MutualAuthChannel() error {
	//not implemented
	return nil
}

//TODO
func (c *MockCard) ListPhonons(limit int, filterType string, filterValue []byte) (phonons []model.Phonon, err error) {
	numStoredPhonons := len(c.Phonons)
	if limit > numStoredPhonons {
		limit = numStoredPhonons
	}
	return phonons[0:limit], nil
}

//TODO
func (c *MockCard) SendPhonons(phononIDs []int) (transaction []byte, err error) {
	//not implemented
	return nil, nil
}

//TODO
func (c *MockCard) ReceivePhonons(transaction []byte) (err error) {
	//not implemented
	return nil
}

//TODO
func (c *MockCard) DestroyPhonon(phononID string) (err error) {
	//not implemented
	return nil
}
