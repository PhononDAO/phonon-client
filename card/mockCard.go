package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"unicode"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
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

//TODO: implement
func (c *MockCard) InitCardPairing() (initPairingData []byte, err error) {
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
