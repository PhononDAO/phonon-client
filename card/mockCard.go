package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"unicode"

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
	cardPubKey, err := NewTLV(TagECCPublicKey, util.SerializeECDSAPubKey(c.IdentityPubKey))
	if err != nil {
		return nil, err
	}
	salt, err := NewTLV(TagSalt, util.RandomKey(32))
	if err != nil {
		return nil, err
	}
	initPairingData = EncodeTLVList(cardCertTLV, cardPubKey, salt)

	return initPairingData, nil
}

func (c *MockCard) CardPair(initCardPairingData []byte) (cardPairingData []byte, err error) {
	tlv, err := ParseTLVPacket(initCardPairingData)
	if err != nil {
		return nil, errors.New("could not parse TLV packet")
	}
	senderCardCertRaw, err := tlv.FindTag(TagCardCertificate)
	if err != nil {
		return nil, errors.New("could not find certificate tlv tag")
	}
	// senderPubKey, err := tlv.FindTag(TagECCPublicKey)
	// if err != nil {
	// 	return nil, errors.New("could not find sender pub key tlv tag")
	// }
	// senderSalt, err := tlv.FindTag(TagSalt)
	// if err != nil {
	// 	return nil, errors.New("could not find sender salt tlv tag")
	// }

	log.Debug("certificate length: ", len(senderCardCertRaw))
	log.Debugf("% X", senderCardCertRaw)
	certLength := senderCardCertRaw[1]
	senderCardCert := CardCertificate{
		Permissions: senderCardCertRaw[2:8],
		PubKey:      senderCardCertRaw[8 : 8+65],
		Sig:         senderCardCertRaw[8+65 : 0+certLength],
	}
	log.Debug("length of Permissions: ", len(senderCardCert.Permissions))
	log.Debugf("Permissions: % X", senderCardCert.Permissions)
	log.Debug("length of PubKey: ", len(senderCardCert.PubKey))
	log.Debugf("PubKey: % X", senderCardCert.PubKey)
	log.Debug("length of Sig: ", len(senderCardCert.Sig))
	log.Debugf("Sig: % X", senderCardCert.Sig)

	valid := ValidateCardCertificate(senderCardCert, gridplus.SafecardDevCAPubKey)
	if !valid {
		return nil, errors.New("counterparty certificate was invalid")
	}
	return nil, nil
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
