package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"unicode"

	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/tlv"
	"github.com/GridPlus/phonon-client/util"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

const StandardSchemaSupportedVersions uint8 = 0

type MockCard struct {
	Phonons []*MockPhonon

	// This is a slice of indeces of deleted phonons. This is to match the insert logic of the card implementation
	deletedPhonons  []int
	pin             string
	pinVerified     bool
	sc              SecureChannel
	receiveList     []*ecdsa.PublicKey
	identityKey     *ecdsa.PrivateKey
	IdentityPubKey  *ecdsa.PublicKey
	IdentityCert    cert.CardCertificate
	scPairData      SecureChannelPairingDetails
	invoices        map[string][]byte
	outgoingInvoice Invoice
	staticPairing   bool
}

type MockPhonon struct {
	model.Phonon
	PrivateKey *ecdsa.PrivateKey
	deleted    bool
}

func (c *MockCard) addPhonon(p *MockPhonon) (index uint16) {
	if len(c.deletedPhonons) > 0 {
		index := c.deletedPhonons[len(c.deletedPhonons)-1]
		c.Phonons[index] = p
		c.deletedPhonons = c.deletedPhonons[:len(c.deletedPhonons)-1]
	} else {
		c.Phonons = append(c.Phonons, p)
		index = uint16(len(c.Phonons) - 1)
	}
	p.KeyIndex = index
	return
}

func (c *MockCard) deletePhonon(index int) {
	c.deletedPhonons = append(c.deletedPhonons, index)
	c.Phonons[index].deleted = true
}

func (phonon *MockPhonon) Encode() (tlv.TLV, error) {
	privKeyTLV, err := tlv.NewTLV(TagPhononPrivKey, phonon.PrivateKey.D.Bytes())
	if err != nil {
		log.Error("could not encode mockPhonon privKey: ", err)
		return tlv.TLV{}, err
	}
	//Also include CurveType
	curveTypeTLV, err := tlv.NewTLV(TagCurveType, []byte{byte(phonon.CurveType)})
	if err != nil {
		return tlv.TLV{}, err
	}
	//Encode internal phonon
	phononTLV, err := TLVEncodePhononDescriptor(&phonon.Phonon)
	if err != nil {
		log.Error("mock could not encode inner phonon: ", phonon.Phonon)
		return tlv.TLV{}, err
	}
	data := append(privKeyTLV.Encode(), curveTypeTLV.Encode()...)
	data = append(data, phononTLV...)
	phononDescriptionTLV, err := tlv.NewTLV(TagPhononPrivateDescription, data)
	if err != nil {
		log.Error("mock could not encode phonon description: ", err)
		return tlv.TLV{}, err
	}

	return phononDescriptionTLV, nil
}

func decodePhononTLV(privatePhononTLV []byte) (phonon MockPhonon, err error) {
	phononTLV, err := tlv.ParseTLVPacket(privatePhononTLV, TagPhononPrivateDescription)
	if err != nil {
		return phonon, err
	}

	//Parse private key for later
	rawPrivKey, err := phononTLV.FindTag(TagPhononPrivKey)
	if err != nil {
		log.Debug("could not parse phonon private key tlv")
		return phonon, err
	}
	//Parse standard public fields and extended schema
	publicPhonon, err := TLVDecodePublicPhononFields(phononTLV)
	if err != nil {
		return phonon, err
	}

	phonon.Phonon = *publicPhonon

	phonon.PrivateKey, err = util.ParseECCPrivKey(rawPrivKey)
	if err != nil {
		return phonon, err
	}
	phonon.PubKey = &phonon.PrivateKey.PublicKey

	return phonon, nil
}

type SecureChannelPairingDetails struct {
	cardToCardSalt     []byte
	counterpartyPubKey *ecdsa.PublicKey
	cryptogram         []byte
	counterPartyCert   cert.CardCertificate
	aesIV              []byte
	encKey             []byte
	macKey             []byte
}

type Invoice struct {
	ID  string //32 length
	Key []byte //32 length
}

func NewMockCard(isInitialized bool, isStatic bool) (*MockCard, error) {
	var identityPrivKey *ecdsa.PrivateKey
	var err error
	if !isStatic {
		identityPrivKey, err = ethcrypto.GenerateKey()
		if err != nil {
			return nil, err
		}
		//If mock is static, use a predetermined identity private key seed
	} else if isStatic {
		var D []byte
		for x := 0; x < 32; x++ {
			D = append(D, 0x01)
		}
		identityPrivKey, err := ethcrypto.ToECDSA(D)
		if err != nil {
			return nil, err
		}
		log.Debugf("generated static privKey: % X\n", identityPrivKey.D.Bytes())
	}

	mockCard := &MockCard{
		identityKey:    identityPrivKey,
		IdentityPubKey: &identityPrivKey.PublicKey,
		invoices:       make(map[string][]byte),
		staticPairing:  isStatic,
	}

	//If card should be initialized, go ahead and install a mock cert and set the test pin
	if isInitialized {
		testPin := "111111"
		mockCard.InstallCertificate(cert.SignWithDemoKey)
		mockCard.Init(testPin)
	}

	return mockCard, nil
}

// //Creates a special debug version of the mock which utilizes static keys and salts
// //to create deterministic payloads for card debugging
// func NewStaticMockCard() (*MockCard, error) {
// 	var D []byte
// 	for x := 0; x < 32; x++ {
// 		D = append(D, 0x01)
// 	}
// 	identityPrivKey, err := ethcrypto.ToECDSA(D)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Debugf("generated static privKey: % X\n", identityPrivKey.D.Bytes())
// 	return &MockCard{
// 		identityKey:    identityPrivKey,
// 		IdentityPubKey: &identityPrivKey.PublicKey,
// 		invoices:       make(map[string][]byte),
// 		staticPairing:  true,
// 	}, nil
// }

// func NewInitializedMockCard() (*MockCard, error) {
// 	mockCard, err := NewMockCard()
// 	if err != nil {
// 		return nil, err
// 	}
// 	testPin := "111111"
// 	mockCard.InstallCertificate(cert.SignWithDemoKey)
// 	mockCard.Init(testPin)
// 	return mockCard, nil
// }

func (c *MockCard) Select() (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error) {
	instanceUID = util.RandomKey(16)

	privKey, _ := ethcrypto.GenerateKey()
	cardPubKey = &privKey.PublicKey

	if c.pin == "" {
		cardInitialized = false
	} else {
		cardInitialized = true
	}
	return instanceUID, cardPubKey, cardInitialized, nil
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
	rawCardCert, err := cert.CreateCardCertificate(c.IdentityPubKey, signKeyFunc)
	if err != nil {
		return err
	}
	log.Debugf("installed cert: % X, len: %v", rawCardCert, len(rawCardCert))
	c.IdentityCert, err = cert.ParseRawCardCertificate(rawCardCert)
	if err != nil {
		return err
	}
	return nil
}

func (c *MockCard) InitCardPairing(receiverCert cert.CardCertificate) (initPairingData []byte, err error) {
	log.Debug("sending mock INIT_CARD_PAIRING command")
	//Ingest counterparty cert and save it for use in CARD_PAIR_2
	log.Debugf("received receiverCert: % X, len: %v", receiverCert.Serialize(), len(receiverCert.Serialize()))
	c.scPairData.counterPartyCert = receiverCert
	_, err = util.ParseECDSAPubKey(receiverCert.PubKey)
	if err != nil {
		return nil, err
	}

	log.Debugf("receiver pubKey: % X", receiverCert.PubKey)
	cardCertTLV, err := tlv.NewTLV(TagCardCertificate, c.IdentityCert.Serialize())
	if err != nil {
		return nil, err
	}
	//Store salt for use in session key generation in CARD_PAIR_2
	if c.staticPairing {
		for x := 0; x < 32; x++ {
			c.scPairData.cardToCardSalt = append(c.scPairData.cardToCardSalt, 0x01)
		}
	} else {
		c.scPairData.cardToCardSalt = util.RandomKey(32)
	}

	log.Debugf("card to card salt: % X\n", c.scPairData.cardToCardSalt)
	saltTLV, err := tlv.NewTLV(TagSalt, c.scPairData.cardToCardSalt)
	if err != nil {
		return nil, err
	}
	initPairingData = tlv.EncodeTLVList(cardCertTLV, saltTLV)

	log.Debugf("returning initPairingData: % X", initPairingData)
	return initPairingData, nil
}

func (c *MockCard) CardPair(initCardPairingData []byte) (cardPairingData []byte, err error) {
	log.Debug("sending mock CARD_PAIR command")
	//Initialize pairing salt
	var receiverSalt []byte
	if c.staticPairing {
		log.Debug("static salting")
		for x := 0; x < 32; x++ {
			receiverSalt = append(receiverSalt, 0x01)
		}
	} else {
		receiverSalt = util.RandomKey(32)
	}
	log.Debugf("receiver salt: % X\n", receiverSalt)

	//Parse Pairing Values from counterparty
	collection, err := tlv.ParseTLVPacket(initCardPairingData)
	if err != nil {
		return nil, errors.New("could not parse TLV packet")
	}
	senderCardCertRaw, err := collection.FindTag(TagCardCertificate)
	if err != nil {
		return nil, errors.New("could not find certificate tlv tag")
	}
	senderSalt, err := collection.FindTag(TagSalt)
	if err != nil {
		return nil, errors.New("could not find sender salt tlv tag")
	}

	senderCardCert, err := cert.ParseRawCardCertificate(senderCardCertRaw)
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
	log.Debugf("Permissions: % X", senderCardCert.Permissions)
	log.Debug("length of PubKey: ", len(senderCardCert.PubKey))
	log.Debugf("PubKey: % X", senderCardCert.PubKey)
	log.Debug("length of Sig: ", len(senderCardCert.Sig))
	log.Debugf("Sig: % X", senderCardCert.Sig)

	//Validate counterparty certificate
	valid := cert.ValidateCardCertificate(senderCardCert, gridplus.SafecardDevCAPubKey)
	if !valid {
		return nil, errors.New("counterparty certificate signature was invalid")
	}

	pubKeyValid := gridplus.ValidateECCPubKey(senderPubKey)
	if !pubKeyValid {
		return nil, errors.New("counterparty public key is not valid ECC point")
	}

	//Compute shared secret
	ecdhSecret := crypto.GenerateECDHSharedSecret(c.identityKey, senderPubKey)
	log.Debugf("sender pubKey: % X", c.IdentityCert.PubKey)
	log.Debugf("ECDH Secret: % X", ecdhSecret)
	log.Debugf("sender salt: % X", senderSalt)
	log.Debugf("receiver salt: % X", receiverSalt)
	//Compute session key with salts from both parties and ECDH secret
	sessionKeyMaterial := append(senderSalt, receiverSalt...)
	sessionKeyMaterial = append(sessionKeyMaterial, ecdhSecret...)

	sessionKey := sha512.Sum512(sessionKeyMaterial)

	log.Debugf("sessionKey: % X\n", sessionKey)
	//Derive secure channel info
	encKey := sessionKey[:len(sessionKey)/2]
	macKey := sessionKey[len(sessionKey)/2:]

	aesIV := util.RandomKey(16)
	c.scPairData.aesIV = make([]byte, len(aesIV))
	c.scPairData.encKey = make([]byte, len(encKey))
	c.scPairData.macKey = make([]byte, len(macKey))
	copy(c.scPairData.aesIV, aesIV)
	copy(c.scPairData.encKey, encKey)
	copy(c.scPairData.macKey, macKey)

	log.Debugf("copied into values: % X, % X, % X", c.scPairData.aesIV, c.scPairData.encKey, c.scPairData.macKey)
	//Combine shared derived session key with randomly generated aesIV and sign to prove possession of the
	//private key corresponding to the public key which established this channel's foundational ECDH secret
	cryptogram := sha256.Sum256(append(sessionKey[0:], aesIV...))
	c.scPairData.cryptogram = cryptogram[0:]
	receiverSig, err := ecdsa.SignASN1(rand.Reader, c.identityKey, cryptogram[0:])
	if err != nil {
		return nil, err
	}

	log.Debugf("cryptogram: % X", cryptogram)
	log.Debugf("receiverSig: % X", receiverSig)
	receiverSaltTLV, _ := tlv.NewTLV(TagSalt, receiverSalt)
	aesIVTLV, _ := tlv.NewTLV(TagAesIV, aesIV)
	receiverSigTLV, _ := tlv.NewTLV(TagECDSASig, receiverSig)

	cardPairingData = append(receiverSaltTLV.Encode(), aesIVTLV.Encode()...)
	cardPairingData = append(cardPairingData, receiverSigTLV.Encode()...)

	return cardPairingData, nil
}

func (c *MockCard) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	log.Debug("sending mock CARD_PAIR_2 command")
	collection, err := tlv.ParseTLVPacket(cardPairData)
	if err != nil {
		return nil, err
	}
	receiverSalt, err := collection.FindTag(TagSalt)
	if err != nil {
		return nil, err
	}
	aesIV, err := collection.FindTag(TagAesIV)
	if err != nil {
		return nil, err
	}
	receiverSig, err := collection.FindTag(TagECDSASig)
	if err != nil {
		return nil, err
	}

	receiverPubKey, err := util.ParseECDSAPubKey(c.scPairData.counterPartyCert.PubKey)
	if err != nil {
		return nil, err
	}
	//Validate counterparty certificate
	valid := cert.ValidateCardCertificate(c.scPairData.counterPartyCert, gridplus.SafecardDevCAPubKey)
	if !valid {
		return nil, errors.New("counterparty certificate signature was invalid")
	}

	pubKeyValid := gridplus.ValidateECCPubKey(receiverPubKey)
	if !pubKeyValid {
		return nil, errors.New("counterparty public key is not valid ECC point")
	}

	//Compute shared secret
	ecdhSecret := crypto.GenerateECDHSharedSecret(c.identityKey, receiverPubKey)

	log.Debugf("ecdh secret: % X", ecdhSecret)
	log.Debugf("sender/cardToCardSalt: % X", c.scPairData.cardToCardSalt)
	log.Debugf("receiverSalt: % X", receiverSalt)
	//Compute session key with salts from both parties and ECDH secret
	sessionKeyMaterial := append(c.scPairData.cardToCardSalt, receiverSalt...)
	sessionKeyMaterial = append(sessionKeyMaterial, ecdhSecret...)

	sessionKey := sha512.Sum512(sessionKeyMaterial)
	log.Debugf("sessionKey: % X\n", sessionKey)

	//Derive secure channel info
	encKey := make([]byte, len(sessionKey)/2)
	macKey := make([]byte, len(sessionKey)/2)
	copy(encKey, sessionKey[:len(sessionKey)/2])
	copy(macKey, sessionKey[len(sessionKey)/2:])

	//Directly initialize instead of using NewSecureChannel() to create secure channel without card channel
	c.sc = SecureChannel{}
	c.sc.Init(aesIV, encKey, macKey)

	//Combine shared derived session key with randomly generated aesIV and sign to prove possession of the
	//private key corresponding to the public key which established this channel's foundational ECDH secret
	cryptogram := sha256.Sum256(append(sessionKey[0:], aesIV...))

	log.Debugf("receiverSig: % X", receiverSig)
	log.Debugf("receiverSig length: %v hex: % X", len(receiverSig), len(receiverSig))

	log.Debugf("receiverPubKey: % X", receiverPubKey)
	//Validate ReceiverSig
	valid = ecdsa.VerifyASN1(receiverPubKey, cryptogram[0:], receiverSig)
	if !valid {
		return nil, errors.New("counterparty cryptogram signature invalid")
	}
	senderSig, err := ecdsa.SignASN1(rand.Reader, c.identityKey, cryptogram[0:])
	if err != nil {
		return nil, err
	}

	senderSigTLV, err := tlv.NewTLV(TagECDSASig, senderSig)
	if err != nil {
		return nil, err
	}

	return senderSigTLV.Encode(), nil
}

func (c *MockCard) FinalizeCardPair(cardPair2Data []byte) (err error) {
	log.Debug("sending mock FINALIZE_CARD_PAIR command")
	tlv, err := tlv.ParseTLVPacket(cardPair2Data)
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

	log.Debugf("initializing channel with values: % X, % X, % X", c.scPairData.aesIV, c.scPairData.encKey, c.scPairData.macKey)
	//Directly initialize instead of using NewSecureChannel() to create secure channel without card channel
	c.sc = SecureChannel{}
	c.sc.Init(c.scPairData.aesIV, c.scPairData.encKey, c.scPairData.macKey)
	return nil
}

func (c *MockCard) Pair() (*cert.CardCertificate, error) {
	//omitted since mockCard does not actually need to establish a secure channel
	return &c.IdentityCert, nil
}

//Phonon Management Functions

func (c *MockCard) CreatePhonon(curveType model.CurveType) (keyIndex uint16, pubKey *ecdsa.PublicKey, err error) {
	if !c.pinVerified {
		return 0, nil, ErrPINNotEntered
	}
	// initialize empty phonon
	newp := MockPhonon{
		deleted: false,
	}
	// generate key
	private, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
	if err != nil {
		return 0, &ecdsa.PublicKey{}, err
	}
	newp.PubKey = &private.PublicKey
	newp.PrivateKey = private
	newp.CurveType = curveType
	//add it in the correct place
	index := c.addPhonon(&newp)

	return index, newp.PubKey, nil
}

func (c *MockCard) SetDescriptor(phonon *model.Phonon) error {
	if int(phonon.KeyIndex) >= len(c.Phonons) || c.Phonons[phonon.KeyIndex].deleted {
		return fmt.Errorf("No phonon at index %d", phonon.KeyIndex)
	}

	storedPhonon := &c.Phonons[phonon.KeyIndex].Phonon

	storedPhonon.SchemaVersion = phonon.SchemaVersion
	storedPhonon.ExtendedSchemaVersion = phonon.ExtendedSchemaVersion
	storedPhonon.CurrencyType = phonon.CurrencyType
	storedPhonon.Denomination = phonon.Denomination
	storedPhonon.ExtendedSchemaVersion = phonon.ExtendedSchemaVersion

	return nil
}

func (c *MockCard) OpenSecureChannel() error {
	//omitted since mockCard does not actually need to establish a secure channel
	return nil
}

func (c *MockCard) ListPhonons(currencyType model.CurrencyType, lessThanValue uint64, greaterThanValue uint64) ([]*model.Phonon, error) {
	var ret []*model.Phonon
	for _, phonon := range c.Phonons {
		if !phonon.deleted &&
			(currencyType == 0x00 || phonon.CurrencyType == currencyType) &&
			(greaterThanValue == 0 || phonon.Denomination.Value() > int(greaterThanValue)) &&
			(lessThanValue == 0 || phonon.Denomination.Value() < int(lessThanValue)) {
			ret = append(ret, &phonon.Phonon)
		}
	}
	return ret, nil
}

func (c *MockCard) GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error) {
	index := int(keyIndex)
	if index > len(c.Phonons) || c.Phonons[index].deleted {
		return &ecdsa.PublicKey{}, fmt.Errorf("No phonon at index %d", index)
	}
	if c.Phonons[index].PubKey == nil {
		return nil, errors.New("phonon pubkey not found. internal error")
	}

	return c.Phonons[index].PubKey, nil
}

func (c *MockCard) SetReceiveList(phononPubKeys []*ecdsa.PublicKey) error {
	c.receiveList = phononPubKeys
	return nil
}

//For invoiced tranfers
// func (c *MockCard) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets []byte, err error) {
// 	var outgoingPhonons []byte
// 	for _, k := range keyIndices {
// 		if int(k) > len(c.Phonons) {
// 			return nil, errors.New("keyIndex exceeds length of phonon list")
// 		}
// 		if c.Phonons[k].deleted {
// 			return nil, errors.New("cannot access deleted phonon")
// 		}
// 		phononTLV, err := c.Phonons[k].Encode()
// 		if err != nil {
// 			return nil, errors.New("could not encode phonon TLV")
// 		}

// 		outgoingPhonons = append(outgoingPhonons, phononTLV.Encode()...)
// 	}
// 	invoiceSC := SecureChannel{}
// 	log.Debugf("invoice before sendPhonon encryption")
// 	log.Debugf("ID: % X", []byte(c.outgoingInvoice.ID))
// 	log.Debugf("Key: % X", c.outgoingInvoice.Key)

// 	//TODO: divide enckey and MAC
// 	invoiceSC.Init([]byte(c.outgoingInvoice.ID), c.outgoingInvoice.Key, c.outgoingInvoice.Key)

// 	phononTransferTLV, err := tlv.NewTLV(TagTransferPhononPacket, outgoingPhonons)
// 	if err != nil {
// 		return nil, errors.New("could not encode phonon transfer TLV")
// 	}

// 	encryptedPhonons, err := invoiceSC.Encrypt(phononTransferTLV.Encode())
// 	if err != nil {
// 		return nil, errors.New("could not encrypt outgoing phonons")
// 	}

// 	invoiceIDTLV, err := tlv.NewTLV(TagInvoiceID, []byte(c.outgoingInvoice.ID))
// 	if err != nil {
// 		return nil, errors.New("could not encode invoice with TLV")
// 	}
// 	response := append(invoiceIDTLV.Encode(), encryptedPhonons...)
// 	return response, nil
// }

func (c *MockCard) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets []byte, err error) {
	log.Debug("mock SEND_PHONONS command")
	var outgoingPhonons []byte
	for _, k := range keyIndices {
		if int(k) >= len(c.Phonons) {
			return nil, errors.New("keyIndex exceeds length of phonon list")
		}
		if c.Phonons[k].deleted {
			return nil, errors.New("cannot access deleted phonon")
		}
		phononTLV, err := c.Phonons[k].Encode()
		if err != nil {
			return nil, errors.New("could not encode phonon TLV")
		}

		outgoingPhonons = append(outgoingPhonons, phononTLV.Encode()...)
	}

	phononTransferTLV, err := tlv.NewTLV(TagTransferPhononPacket, outgoingPhonons)
	if err != nil {
		return nil, errors.New("could not encode phonon transfer TLV")
	}

	encryptedPhonons, err := c.sc.Encrypt(phononTransferTLV.Encode())
	if err != nil {
		return nil, errors.New("could not encrypt outgoing phonons")
	}

	//Delete sent phonons
	for _, k := range keyIndices {
		c.deletePhonon(int(k))
	}

	return encryptedPhonons, nil
}

//For invoiced receives
// func (c *MockCard) ReceivePhonons(transaction []byte) (err error) {
// 	data, err := tlv.ParseTLVPacket(transaction)
// 	if err != nil {
// 		return err
// 	}
// 	invoiceID, err := data.FindTag(TagInvoiceID)
// 	if err != nil {
// 		return err
// 	}
// 	encKey, ok := c.invoices[string(invoiceID)]
// 	if !ok {
// 		return errors.New("invoiceID not found")
// 	}
// 	delete(c.invoices, string(invoiceID))

// 	//Grab the encrypted data after the 2 byte TLV + invoiceID
// 	encData := transaction[len(invoiceID)+2:]

// 	receiveSC := SecureChannel{}
// 	receiveSC.Init(invoiceID, encKey, encKey)
// 	phononTransferPacketData, err := receiveSC.Decrypt(encData)
// 	if err != nil {
// 		return err
// 	}

// 	phononTransferPacketTLV, err := tlv.ParseTLVPacket(phononTransferPacketData, TagTransferPhononPacket)
// 	if err != nil {
// 		return err
// 	}

// 	phononTLVs, err := phononTransferPacketTLV.FindTags(TagPhononPrivateDescription)
// 	if err != nil {
// 		return err
// 	}

// 	//Parse all received phonons
// 	var phonons []MockPhonon
// 	for _, tlv := range phononTLVs {
// 		phonon, err := decodePhononTLV(tlv)
// 		if err != nil {
// 			return err
// 		}
// 		phonons = append(phonons, phonon)
// 	}
// 	//Store all received phonons
// 	for _, p := range phonons {
// 		log.Debugf("adding phonon to table: %+v", p)
// 		c.addPhonon(p)
// 	}

// 	return nil
// }

func (c *MockCard) ReceivePhonons(transaction []byte) (err error) {
	log.Debug("mock RECEIVE_PHONONS command")
	phononTransferPacketData, err := c.sc.Decrypt(transaction)
	if err != nil {
		log.Debug("error decrypting incoming phonon transfer packet")
		return err
	}

	phononTransferPacketTLV, err := tlv.ParseTLVPacket(phononTransferPacketData, TagTransferPhononPacket)
	if err != nil {
		return err
	}

	phononTLVs, err := phononTransferPacketTLV.FindTags(TagPhononPrivateDescription)
	if err != nil {
		return err
	}

	//Parse all received phonons
	var phonons []MockPhonon
	for _, tlv := range phononTLVs {
		phonon, err := decodePhononTLV(tlv)
		if err != nil {
			return err
		}
		phonons = append(phonons, phonon)
	}
	//Store all received phonons
	for _, p := range phonons {
		c.addPhonon(&p)
	}

	return nil
}

func (c *MockCard) TransactionAck(keyIndices []uint16) error {
	return nil
}

func (c *MockCard) DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error) {
	index := int(keyIndex)
	c.deletedPhonons = append(c.deletedPhonons, index)
	c.Phonons[index].deleted = true
	return c.Phonons[index].PrivateKey, nil
}

func (c *MockCard) GenerateInvoice() (invoiceData []byte, err error) {
	invoiceID := string(util.RandomKey(16))
	invoiceKey := util.RandomKey(32)

	c.invoices[invoiceID] = invoiceKey

	keyTLV, err := tlv.NewTLV(TagAESKey, invoiceKey)
	if err != nil {
		return nil, err
	}
	idTLV, err := tlv.NewTLV(TagAesIV, []byte(invoiceID))
	if err != nil {
		return nil, err
	}
	data := append(keyTLV.Encode(), idTLV.Encode()...)

	encData, err := c.sc.Encrypt(data)
	if err != nil {
		return nil, err
	}

	return encData, nil
}

func (c *MockCard) ReceiveInvoice(invoiceData []byte) (err error) {
	data, err := c.sc.Decrypt(invoiceData)
	if err != nil {
		return err
	}
	collection, err := tlv.ParseTLVPacket(data)
	if err != nil {
		return err
	}
	invoiceKey, err := collection.FindTag(TagAESKey)
	if err != nil {
		return err
	}
	invoiceID, err := collection.FindTag(TagAesIV)
	if err != nil {
		return err
	}

	//One invoice active at a time
	c.outgoingInvoice = Invoice{
		ID:  string(invoiceID),
		Key: invoiceKey,
	}
	log.Debugf("mock setting outgoingInvoice ID: % X, Key: % X", c.outgoingInvoice.ID, c.outgoingInvoice.Key)
	return nil
}
