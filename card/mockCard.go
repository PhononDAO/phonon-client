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
	scPairData     SecureChannelPairingDetails
}

type SecureChannelPairingDetails struct {
	cardToCardSalt     []byte
	counterpartyPubKey *ecdsa.PublicKey
	cryptogram         []byte
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
	log.Debug("sending mock INIT_CARD_PAIRING command")
	cardCertTLV, err := NewTLV(TagCardCertificate, c.IdentityCert)
	if err != nil {
		return nil, err
	}
	salt, err := NewTLV(TagSalt, util.RandomKey(32))
	if err != nil {
		return nil, err
	}
	//Store salt for use in session key generation in CARD_PAIR_2
	c.scPairData.cardToCardSalt = salt.value
	initPairingData = EncodeTLVList(cardCertTLV, salt)

	return initPairingData, nil
}

func (c *MockCard) CardPair(initCardPairingData []byte) (cardPairingData []byte, err error) {
	log.Debug("sending mock CARD_PAIR command")
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
	//Store sender's public key for signature validation in FINALIZE_CARD_PAIR
	c.scPairData.counterpartyPubKey = senderPubKey

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
	encKey := sessionKey[:len(sessionKey)/2]
	macKey := sessionKey[len(sessionKey)/2:]

	aesIV := util.RandomKey(16)

	//Directly initialize instead of using NewSecureChannel() to create secure channel without card channel
	c.sc = SecureChannel{}
	c.sc.Init(aesIV, encKey, macKey)

	//Combine shared derived session key with randomly generated aesIV and sign to prove possession of the
	//private key corresponding to the public key which established this channel's foundational ECDH secret
	cryptogram := sha256.Sum256(append(sessionKey[0:], aesIV...))
	c.scPairData.cryptogram = cryptogram[0:]
	receiverSig, err := ecdsa.SignASN1(rand.Reader, c.identityKey, cryptogram[0:])
	if err != nil {
		return nil, err
	}

	receiverCertTLV, _ := NewTLV(TagCardCertificate, c.IdentityCert)
	receiverSaltTLV, _ := NewTLV(TagSalt, receiverSalt)
	aesIVTLV, _ := NewTLV(TagAesIV, aesIV)
	receiverSigTLV, _ := NewTLV(TagECDSASig, receiverSig)

	cardPairingData = append(receiverCertTLV.Encode(), receiverSaltTLV.Encode()...)
	cardPairingData = append(cardPairingData, aesIVTLV.Encode()...)
	cardPairingData = append(cardPairingData, receiverSigTLV.Encode()...)

	return cardPairingData, nil
}

func (c *MockCard) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	log.Debug("sending mock CARD_PAIR_2 command")
	tlv, err := ParseTLVPacket(cardPairData)
	if err != nil {
		return nil, err
	}
	receiverCardCertRaw, err := tlv.FindTag(TagCardCertificate)
	if err != nil {
		return nil, err
	}
	receiverSalt, err := tlv.FindTag(TagSalt)
	if err != nil {
		return nil, err
	}
	aesIV, err := tlv.FindTag(TagAesIV)
	if err != nil {
		return nil, err
	}
	receiverSig, err := tlv.FindTag(TagECDSASig)
	if err != nil {
		return nil, err
	}

	//Mirror of other side's CARD_PAIR
	receiverCardCert, err := ParseRawCardCertificate(receiverCardCertRaw)
	if err != nil {
		return nil, err
	}
	receiverPubKey, err := util.ParseECDSAPubKey(receiverCardCert.PubKey)
	if err != nil {
		return nil, err
	}
	//Validate counterparty certificate
	valid := ValidateCardCertificate(receiverCardCert, gridplus.SafecardDevCAPubKey)
	if !valid {
		return nil, errors.New("counterparty certificate signature was invalid")
	}

	pubKeyValid := gridplus.ValidateECCPubKey(receiverPubKey)
	if !pubKeyValid {
		return nil, errors.New("counterparty public key is not valid ECC point")
	}

	//Compute shared secret
	ecdhSecret := crypto.GenerateECDHSharedSecret(c.identityKey, receiverPubKey)

	//Compute session key with salts from both parties and ECDH secret
	sessionKeyMaterial := append(c.scPairData.cardToCardSalt, receiverSalt...)
	sessionKeyMaterial = append(sessionKeyMaterial, ecdhSecret...)

	sessionKey := sha512.Sum512(sessionKeyMaterial)

	//Derive secure channel info
	encKey := sessionKey[:len(sessionKey)/2]
	macKey := sessionKey[len(sessionKey)/2:]

	//Directly initialize instead of using NewSecureChannel() to create secure channel without card channel
	c.sc = SecureChannel{}
	c.sc.Init(aesIV, encKey, macKey)

	//Combine shared derived session key with randomly generated aesIV and sign to prove possession of the
	//private key corresponding to the public key which established this channel's foundational ECDH secret
	cryptogram := sha256.Sum256(append(sessionKey[0:], aesIV...))

	//Validate ReceiverSig
	valid = ecdsa.VerifyASN1(receiverPubKey, cryptogram[0:], receiverSig)
	if !valid {
		return nil, errors.New("counterparty cryptogram signature invalid")
	}
	senderSig, err := ecdsa.SignASN1(rand.Reader, c.identityKey, cryptogram[0:])
	if err != nil {
		return nil, err
	}

	senderSigTLV, err := NewTLV(TagECDSASig, senderSig)
	if err != nil {
		return nil, err
	}

	return senderSigTLV.Encode(), nil
}

func (c *MockCard) FinalizeCardPair(cardPair2Data []byte) (err error) {
	log.Debug("sending mock FINALIZE_CARD_PAIR command")
	tlv, err := ParseTLVPacket(cardPair2Data)
	if err != nil {
		return err
	}
	senderSig, err := tlv.FindTag(TagECDSASig)
	if err != nil {
		return err
	}
	//Validate SenderSig
	valid := ecdsa.VerifyASN1(c.scPairData.counterpartyPubKey, c.scPairData.cryptogram, senderSig)
	if !valid {
		return errors.New("counterparty cryptogram signature invalid")
	}
	return nil
}

func (c *MockCard) Pair() error {
	//TODO
	return nil
}

//Phonon Management Functions
//TODO
func (c *MockCard) CreatePhonon() (keyIndex uint16, pubKey *ecdsa.PublicKey, err error) {
	//TODO
	return 0, &ecdsa.PublicKey{}, nil
}

//TODO
func (c *MockCard) SetDescriptor(keyIndex uint16, currencyType model.CurrencyType, value float32) error {
	//TODO
	return nil
}

func (c *MockCard) OpenSecureChannel() error {
	//TODO
	return nil
}

//TODO
func (c *MockCard) ListPhonons(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) ([]model.Phonon, error) {
	//TODO
	return nil, nil
}

func (c *MockCard) GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error) {
	//TODO
	return &ecdsa.PublicKey{}, nil
}

func (c *MockCard) SetReceiveList(phononPubKeys []*ecdsa.PublicKey) error {
	//TODO
	return nil
}

//TODO
func (c *MockCard) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets [][]byte, err error) {
	return nil, nil
}

//TODO
func (c *MockCard) ReceivePhonons(transaction []byte) (err error) {
	//not implemented
	return nil
}

func (c *MockCard) TransactionAck(keyIndices []uint16) error {
	return nil
}

//TODO
func (c *MockCard) DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error) {
	return &ecdsa.PrivateKey{}, nil
}
