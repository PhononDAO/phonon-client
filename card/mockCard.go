package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode"

	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type MockCard struct {
	Phonons []MockPhonon

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
}

type MockPhonon struct {
	model.Phonon
	PrivateKey *ecdsa.PrivateKey
	deleted    bool
}

func (c *MockCard) addPhonon(p MockPhonon) {
	var index uint16
	if len(c.deletedPhonons) > 0 {
		index := c.deletedPhonons[len(c.deletedPhonons)-1]
		c.Phonons[index] = p
		c.deletedPhonons = c.deletedPhonons[:len(c.deletedPhonons)-1]
	} else {
		c.Phonons = append(c.Phonons, p)
		index = uint16(len(c.Phonons) - 1)

	}
	p.KeyIndex = index
}

func (phonon *MockPhonon) Encode() (TLV, error) {
	privKeyTLV, _ := NewTLV(TagPhononPrivKey, phonon.PrivateKey.D.Bytes())
	valueBytes, _ := util.Float32ToBytes(phonon.Value)
	valueTLV, _ := NewTLV(TagPhononValue, valueBytes)

	currencyTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyTypeBytes, uint16(phonon.CurrencyType))

	currencyTypeTLV, _ := NewTLV(TagCurrencyType, currencyTypeBytes)

	phononDescriptionBytes := append(privKeyTLV.Encode(), valueTLV.Encode()...)
	phononDescriptionBytes = append(phononDescriptionBytes, currencyTypeTLV.Encode()...)

	phononDescriptionTLV, _ := NewTLV(TagPhononPrivateDescription, phononDescriptionBytes)

	return phononDescriptionTLV, nil
}

func decodePhononTLV(privatePhononTLV []byte) (phonon MockPhonon, err error) {
	phononTLV, err := ParseTLVPacket(privatePhononTLV, TagPhononPrivateDescription)
	if err != nil {
		return phonon, err
	}
	//privKey
	rawPrivKey, err := phononTLV.FindTag(TagPhononPrivKey)
	if err != nil {
		return phonon, err
	}
	log.Debugf("rawPrivKeyBytes: % X", rawPrivKey)
	phonon.PrivateKey, err = util.ParseECCPrivKey(rawPrivKey)
	if err != nil {
		return phonon, err
	}
	phonon.PubKey = &phonon.PrivateKey.PublicKey

	//value
	valueBytes, err := phononTLV.FindTag(TagPhononValue)
	if err != nil {
		return phonon, err
	}
	phonon.Value, err = util.BytesToFloat32(valueBytes)
	if err != nil {
		return phonon, err
	}

	//currencyType
	currencyTypeBytes, err := phononTLV.FindTag(TagCurrencyType)
	if err != nil {
		return phonon, err
	}
	log.Debugf("currencyTypeBytes: % X", currencyTypeBytes)
	phonon.CurrencyType = model.CurrencyType(binary.BigEndian.Uint16(currencyTypeBytes))

	return phonon, nil
}

type SecureChannelPairingDetails struct {
	cardToCardSalt     []byte
	counterpartyPubKey *ecdsa.PublicKey
	cryptogram         []byte
	counterPartyCert   cert.CardCertificate
}

type Invoice struct {
	ID  string //32 length
	Key []byte //32 length
}

func NewMockCard() (*MockCard, error) {
	identityPrivKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	return &MockCard{
		identityKey:    identityPrivKey,
		IdentityPubKey: &identityPrivKey.PublicKey,
		invoices:       make(map[string][]byte),
	}, nil
}

func (c *MockCard) Select() (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error) {
	instanceUID = util.RandomKey(16)

	privKey, _ := ethcrypto.GenerateKey()
	cardPubKey = &privKey.PublicKey

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
	rawCardCert, err := cert.CreateCardCertificate(c.IdentityPubKey, signKeyFunc)
	c.IdentityCert, err = cert.ParseRawCardCertificate(rawCardCert)
	if err != nil {
		return err
	}
	return nil
}

func (c *MockCard) InitCardPairing(receiverCert cert.CardCertificate) (initPairingData []byte, err error) {
	log.Debug("sending mock INIT_CARD_PAIRING command")
	//Ingest coutnerparty cert and save it for use in CARD_PAIR_2
	c.scPairData.counterPartyCert = receiverCert

	cardCertTLV, err := NewTLV(TagCardCertificate, c.IdentityCert.Serialize())
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

	receiverCertTLV, _ := NewTLV(TagCardCertificate, c.IdentityCert.Serialize())
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
	// receiverCardCertRaw, err := tlv.FindTag(Tagcert.CardCertificate)
	// if err != nil {
	// 	return nil, err
	// }
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

	// //Mirror of other side's CARD_PAIR
	// receiverCardCert, err := ParseRawcert.CardCertificate(receiverCardCertRaw)
	// if err != nil {
	// 	return nil, err
	// }

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

func (c *MockCard) Pair() (cert.CardCertificate, error) {
	//TODO
	return c.IdentityCert, nil
}

//Phonon Management Functions

func (c *MockCard) CreatePhonon() (keyIndex uint16, pubKey *ecdsa.PublicKey, err error) {
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
	var index int16
	//add it in the correct place
	c.addPhonon(newp)

	return uint16(index), newp.PubKey, nil
}

func (c *MockCard) SetDescriptor(keyIndex uint16, currencyType model.CurrencyType, value float32) error {
	index := int(keyIndex)
	if index > len(c.Phonons) || c.Phonons[index].deleted {
		return fmt.Errorf("No phonon at index %d", index)
	}
	c.Phonons[index].CurrencyType = currencyType
	c.Phonons[index].Value = value
	return nil
}

func (c *MockCard) OpenSecureChannel() error {
	//TODO
	return nil
}

func (c *MockCard) ListPhonons(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) ([]model.Phonon, error) {
	var ret []model.Phonon
	for _, phonon := range c.Phonons {
		if !phonon.deleted &&
			(currencyType == 0x00 || phonon.CurrencyType == currencyType) &&
			phonon.Value > greaterThanValue &&
			(lessThanValue == 0 || phonon.Value < lessThanValue) {
			ret = append(ret, phonon.Phonon)
		}
	}
	return ret, nil
}

func (c *MockCard) GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error) {
	index := int(keyIndex)
	if index > len(c.Phonons) || c.Phonons[index].deleted {
		return &ecdsa.PublicKey{}, fmt.Errorf("No phonon at index %d", index)
	}
	return c.Phonons[index].PubKey, nil
}

func (c *MockCard) SetReceiveList(phononPubKeys []*ecdsa.PublicKey) error {
	c.receiveList = phononPubKeys
	return nil
}

func (c *MockCard) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets []byte, err error) {
	var outgoingPhonons []byte
	for _, k := range keyIndices {
		if int(k) > len(c.Phonons) {
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
	invoiceSC := SecureChannel{}
	log.Debugf("invoice before sendPhonon encryption")
	log.Debugf("ID: % X", []byte(c.outgoingInvoice.ID))
	log.Debugf("Key: % X", c.outgoingInvoice.Key)

	//TODO: divide enckey and MAC
	invoiceSC.Init([]byte(c.outgoingInvoice.ID), c.outgoingInvoice.Key, c.outgoingInvoice.Key)

	phononTransferTLV, err := NewTLV(TagTransferPhononPacket, outgoingPhonons)
	if err != nil {
		return nil, errors.New("could not encode phonon transfer TLV")
	}

	encryptedPhonons, err := invoiceSC.Encrypt(phononTransferTLV.Encode())
	if err != nil {
		return nil, errors.New("could not encrypt outgoing phonons")
	}

	invoiceIDTLV, err := NewTLV(TagInvoiceID, []byte(c.outgoingInvoice.ID))
	if err != nil {
		return nil, errors.New("could not encode invoice with TLV")
	}
	response := append(invoiceIDTLV.Encode(), encryptedPhonons...)
	return response, nil
}

func (c *MockCard) ReceivePhonons(transaction []byte) (err error) {
	data, err := ParseTLVPacket(transaction)
	if err != nil {
		return err
	}
	invoiceID, err := data.FindTag(TagInvoiceID)
	if err != nil {
		return err
	}
	encKey, ok := c.invoices[string(invoiceID)]
	if !ok {
		return errors.New("invoiceID not found")
	}
	delete(c.invoices, string(invoiceID))

	//Grab the encrypted data after the 2 byte TLV + invoiceID
	encData := transaction[len(invoiceID)+2:]

	receiveSC := SecureChannel{}
	receiveSC.Init(invoiceID, encKey, encKey)
	phononTransferPacketData, err := receiveSC.Decrypt(encData)
	if err != nil {
		return err
	}

	phononTransferPacketTLV, err := ParseTLVPacket(phononTransferPacketData, TagTransferPhononPacket)
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
		log.Debugf("adding phonon to table: %+v", p)
		c.addPhonon(p)
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

	keyTLV, err := NewTLV(TagAESKey, invoiceKey)
	if err != nil {
		return nil, err
	}
	idTLV, err := NewTLV(TagAesIV, []byte(invoiceID))
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
	collection, err := ParseTLVPacket(data)
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
