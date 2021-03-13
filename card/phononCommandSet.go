package card

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/GridPlus/keycard-go"
	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/globalplatform"
	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/keycard-go/types"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

var phononAID = []byte{0xA0, 0x00, 0x00, 0x08, 0x20, 0x00, 0x03, 0x01}

type PhononCommandSet struct {
	c               types.Channel
	sc              *SecureChannel
	ApplicationInfo *types.ApplicationInfo //TODO: Determine if needed
	PairingInfo     *types.PairingInfo
}

func NewPhononCommandSet(c types.Channel) *PhononCommandSet {
	return &PhononCommandSet{
		c:               c,
		sc:              NewSecureChannel(c),
		ApplicationInfo: &types.ApplicationInfo{},
	}
}

//TODO: determine if I should return these values or have the secure channel handle it internally
//Selects the phonon applet for further usage
//Returns ErrCardUninitialized if card is not yet initialized with a pin
func (cs *PhononCommandSet) Select() (instanceUID []byte, cardPubKey []byte, err error) {
	cmd := globalplatform.NewCommandSelect(phononAID)
	cmd.SetLe(0)

	log.Debug("sending SELECT apdu")
	resp, err := cs.c.Send(cmd)
	if err != nil {
		log.Error("could not send select command. err: ", err)
		return nil, nil, err
	}

	instanceUID, cardPubKey, err = ParseSelectResponse(resp.Data) //unused var is intanceUID
	if err != nil && err != ErrCardUninitialized {
		log.Error("error parsing select response. err: ", err)
		return nil, nil, err
	}

	//TODO: Use random version GenerateSecret in production
	//Generate secure channel secrets using card's public key
	secretsErr := cs.sc.GenerateSecret(cardPubKey)
	if secretsErr != nil {
		log.Error("could not generate secure channel secrets. err: ", secretsErr)
		return nil, nil, secretsErr
	}
	log.Debug("Pairing generated key:\n", hex.Dump(cs.sc.RawPublicKey()))
	//return ErrCardUninitialized if ParseSelectResponse returns that error code
	return instanceUID, cardPubKey, err
}

func (cs *PhononCommandSet) Pair() error {
	//Generate random salt and keypair
	clientSalt := make([]byte, 32)
	rand.Read(clientSalt)

	pairingPrivKey, err := ethcrypto.GenerateKey()
	if err != nil {
		log.Error("unable to generate pairing keypair. err: ", err)
		return err
	}
	pairingPubKey := pairingPrivKey.PublicKey

	//Exchange pairing key info with card
	cmd := gridplus.NewAPDUPairStep1(clientSalt, &pairingPubKey)
	resp, err := cs.c.Send(cmd)
	if err != nil {
		log.Error("unable to send Pair Step 1 command. err: ", err)
		return err
	}
	pairStep1Resp, err := gridplus.ParsePairStep1Response(resp.Data)
	if err != nil {
		log.Error("could not parse pair step 2 response. err: ", err)
	}

	//Validate card's certificate has valid GridPlus signature
	certValid := gridplus.ValidateCardCertificate(pairStep1Resp.SafecardCert)
	log.Debug("certificate signature valid: ", certValid)
	if !certValid {
		log.Error("unable to verify card certificate.")
		return err
	}
	log.Debug("pair step 2 safecard cert:\n", hex.Dump(pairStep1Resp.SafecardCert.PubKey))

	cardCertPubKey, err := gridplus.ParseCertPubkeyToECDSA(pairStep1Resp.SafecardCert.PubKey)
	if err != nil {
		log.Error("unable to parse certificate public key. err: ", err)
		return err
	}

	pubKeyValid := gridplus.ValidateECCPubKey(cardCertPubKey)
	log.Debug("certificate public key valid: ", pubKeyValid)
	if !pubKeyValid {
		log.Error("card pubkey invalid")
		return err
	}

	//challenge message test
	ecdhSecret := crypto.GenerateECDHSharedSecret(pairingPrivKey, cardCertPubKey)

	secretHashArray := sha256.Sum256(append(clientSalt, ecdhSecret...))
	secretHash := secretHashArray[0:]

	type ECDSASignature struct {
		R, S *big.Int
	}
	signature := &ECDSASignature{}
	_, err = asn1.Unmarshal(pairStep1Resp.SafecardSig, signature)
	if err != nil {
		log.Error("could not unmarshal certificate signature.", err)
	}

	//validate that card created valid signature over same salted and hashed ecdh secret
	valid := ecdsa.Verify(cardCertPubKey, secretHash, signature.R, signature.S)
	if !valid {
		log.Error("ecdsa sig not valid")
		return errors.New("could not verify shared secret challenge")
	}
	log.Debug("card signature on challenge message valid: ", valid)

	cryptogram := sha256.Sum256(append(pairStep1Resp.SafecardSalt, secretHash...))

	log.Debug("sending pair step 2 cmd")
	cmd = gridplus.NewAPDUPairStep2(cryptogram[0:])
	resp, err = cs.c.Send(cmd)
	if err != nil {
		log.Error("error sending pair step 2 command. err: ", err)
		return err
	}

	pairStep2Resp, err := gridplus.ParsePairStep2Response(resp.Data)
	if err != nil {
		log.Error("could not parse pair step 2 response. err: ", err)
	}
	log.Debugf("pairStep2Resp: % X", pairStep2Resp)

	//Derive Pairing Key
	pairingKey := sha256.Sum256(append(pairStep2Resp.Salt, secretHash...))
	log.Debugf("derived pairing key: % X", pairingKey)

	//Store pairing info for use in OpenSecureChannel
	cs.SetPairingInfo(pairingKey[0:], pairStep2Resp.PairingIdx)

	log.Debug("pairing succeeded")
	return nil
}

func (cs *PhononCommandSet) SetPairingInfo(key []byte, index int) {
	cs.PairingInfo = &types.PairingInfo{
		Key:   key,
		Index: index,
	}
}

func (cs *PhononCommandSet) Unpair(index uint8) error {
	cmd := keycard.NewCommandUnpair(index)
	resp, err := cs.sc.Send(cmd)
	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) OpenSecureChannel() error {
	if cs.ApplicationInfo == nil {
		return errors.New("cannot open secure channel without setting PairingInfo")
	}

	cmd := keycard.NewCommandOpenSecureChannel(uint8(cs.PairingInfo.Index), cs.sc.RawPublicKey())
	resp, err := cs.c.Send(cmd)
	if err = cs.checkOK(resp, err); err != nil {
		return err
	}

	encKey, macKey, iv := crypto.DeriveSessionKeys(cs.sc.Secret(), cs.PairingInfo.Key, resp.Data)
	cs.sc.Init(iv, encKey, macKey)

	err = cs.mutualAuthenticate()
	if err != nil {
		return err
	}

	return nil
}

func (cs *PhononCommandSet) mutualAuthenticate() error {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return err
	}

	cmd := keycard.NewCommandMutuallyAuthenticate(data)
	resp, err := cs.sc.Send(cmd)

	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) Init(pin string) error {
	log.Debug("sending INIT apdu")
	secrets, err := keycard.GenerateSecrets()
	if err != nil {
		log.Error("unable to generate secrets: ", err)
	}

	//Reusing keycard Secrets implementation with PUK removed for now.
	data, err := crypto.OneShotEncrypt(cs.sc.RawPublicKey(), cs.sc.Secret(), append([]byte(pin), secrets.PairingToken()...))
	if err != nil {
		return err
	}
	log.Debug("len of data: ", len(data))
	init := keycard.NewCommandInit(data)
	resp, err := cs.c.Send(init)

	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) checkOK(resp *apdu.Response, err error, allowedResponses ...uint16) error {
	if err != nil {
		return err
	}

	if len(allowedResponses) == 0 {
		allowedResponses = []uint16{apdu.SwOK}
	}

	for _, code := range allowedResponses {
		if code == resp.Sw {
			return nil
		}
	}

	return apdu.NewErrBadResponse(resp.Sw, "unexpected response")
}

func (cs *PhononCommandSet) IdentifyCard(nonce []byte) (cardPubKey []byte, cardSig []byte, err error) {
	cmd := NewCommandIdentifyCard(nonce)
	resp, err := cs.c.Send(cmd)
	if err != nil {
		log.Error("could not send identify card command", err)
		return nil, nil, err
	}
	log.Debug("identify card resp:\n", hex.Dump(resp.Data))

	cardPubKey, cardSig, err = ParseIdentifyCardResponse(resp.Data)
	if err != nil {
		log.Error("could not parse identify card response: ", err)
		return nil, nil, err
	}

	return cardPubKey, cardSig, nil
}

func (cs *PhononCommandSet) VerifyPIN(pin string) error {
	cmd := NewCommandVerifyPIN(pin)
	resp, err := cs.sc.Send(cmd)
	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) ChangePIN(pin string) error {
	cmd := NewCommandChangePIN(pin)
	resp, err := cs.sc.Send(cmd)

	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) CreatePhonon() (keyIndex int, pubKey *ecdsa.PublicKey, err error) {
	cmd := NewCommandCreatePhonon()
	log.Info("sending create phonon command")
	resp, err := cs.c.Send(cmd) //temp normal channel for testing
	if err != nil {
		log.Error("create phonon command failed: ", err)
		return 0, nil, err
	}
	if resp.Sw == StatusPhononTableFull {
		return 0, nil, ErrPhononTableFull
	}
	keyIndex, pubKey, err = ParseCreatePhononResponse(resp.Data)
	if err != nil {
		return 0, nil, err
	}
	return keyIndex, pubKey, nil
}

func (cs *PhononCommandSet) SetDescriptor(keyIndex uint16, currencyType model.CurrencyType, value float32) error {
	data, err := encodeSetDescriptorData(keyIndex, currencyType, value)
	if err != nil {
		return err
	}

	cmd := NewCommandSetDescriptor(data)
	resp, err := cs.c.Send(cmd) //temp normal channel for testing
	if err != nil {
		log.Error("set descriptor command failed: ", err)
		return err
	}

	return cs.checkOK(resp, err)
}

func encodeSetDescriptorData(keyIndex uint16, currencyType model.CurrencyType, value float32) ([]byte, error) {
	keyIndexBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(keyIndexBytes, keyIndex)
	keyIndexTLV, err := NewTLV(TagKeyIndex, keyIndexBytes)
	if err != nil {
		return nil, err
	}

	currencyBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyBytes, uint16(currencyType))
	currencyTypeTLV, err := NewTLV(TagCurrencyType, currencyBytes)
	if err != nil {
		return nil, err
	}

	var valueBytes bytes.Buffer
	err = binary.Write(&valueBytes, binary.BigEndian, value)
	if err != nil {
		log.Error("unable to write float value as bytes: ", err)
		return nil, err
	}
	valueTLV, err := NewTLV(TagPhononValue, valueBytes.Bytes())
	if err != nil {
		return nil, err
	}

	descriptorBytes := append(keyIndexTLV.Encode(), currencyTypeTLV.Encode()...)
	descriptorBytes = append(descriptorBytes, valueTLV.Encode()...)
	phononDescriptorTLV, err := NewTLV(TagPhononDescriptor, descriptorBytes)
	if err != nil {
		return nil, err
	}
	return phononDescriptorTLV.Encode(), nil
}

func (cs *PhononCommandSet) ListPhonons(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) ([]model.Phonon, error) {
	log.Debug("sending list phonons command")
	p2, cmdData, err := encodeListPhononsData(currencyType, lessThanValue, greaterThanValue)
	if err != nil {
		return nil, err
	}
	log.Debug("List phonons command data: ")
	log.Debugf("% X", cmdData)
	log.Debugf("p2: % X", p2)
	cmd := NewCommandListPhonons(0x00, p2, cmdData)
	resp, err := cs.c.Send(cmd)
	if err != nil {
		return nil, err
	}

	continues, err := checkStatusWord(resp.Sw)
	if err != nil {
		return nil, err
	}

	phonons, err := parseListPhononsResponse(resp.Data)
	if err != nil {
		log.Error("could not parse list phonons response: ", err)
		return nil, err
	}
	if continues {
		extendedPhonons, err := cs.listPhononsExtended()
		if err != nil {
			log.Error("could not read extended phonons list: ", err)
			return nil, err
		}
		phonons = append(phonons, extendedPhonons...)
	}
	return phonons, nil
}

//Makes an additional list phonons command with p1 set to 0x01, indicating the card should return the remainder
//of the last requested list. listPhononsExtended will run recursively until the card indicates there are no additional
//phonons in the list
func (cs *PhononCommandSet) listPhononsExtended() (phonons []model.Phonon, err error) {
	log.Debug("sending request for extended phonons list")
	cmd := NewCommandListPhonons(0x01, 0x00, nil)
	resp, err := cs.c.Send(cmd)
	if err != nil {
		return nil, err
	}
	continues, err := checkStatusWord(resp.Sw)
	if err != nil {
		return nil, err
	}

	phonons, err = parseListPhononsResponse(resp.Data)
	if err != nil {
		log.Error("could not parse extended list phonons response: ", err)
		return nil, err
	}

	if continues {
		extendedPhonons, err := cs.listPhononsExtended()
		if err != nil {
			log.Error("could not read additional extendend phonons list: ", err)
			return nil, err
		}
		phonons = append(phonons, extendedPhonons...)
	}
	return phonons, nil
}

func encodeListPhononsData(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) (p2 byte, data []byte, err error) {
	//Toggle filter bytes for nonzero lessThan and greaterThan filter values
	if lessThanValue == 0 {
		//Don't filter on value at all
		if greaterThanValue == 0 {
			p2 = 0x00
		}
		//Filter on only GreaterThan Value
		if greaterThanValue > 0 {
			p2 = 0x02
		}
	}
	if lessThanValue > 0 {
		//Filter on only LessThanValue
		if greaterThanValue == 0 {
			p2 = 0x01
		}
		//Filter on LessThan and GreaterThan
		if greaterThanValue > 0 {
			p2 = 0x03
		}

	}

	//Translate currencyType to bytes
	currencyTypeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(currencyTypeBytes, uint16(currencyType))

	currencyTypeTLV, err := NewTLV(TagCurrencyType, currencyTypeBytes)
	if err != nil {
		return p2, nil, err
	}
	//Translate filter values to bytes
	lessThanBytes, err := util.Float32ToBytes(lessThanValue)
	if err != nil {
		return p2, nil, err
	}
	greaterThanBytes, err := util.Float32ToBytes(greaterThanValue)
	if err != nil {
		return p2, nil, err
	}
	lessThanTLV, err := NewTLV(TagValueFilterLessThan, lessThanBytes)
	if err != nil {
		return p2, nil, err
	}
	greaterThanTLV, err := NewTLV(TagValueFilterMoreThan, greaterThanBytes)
	if err != nil {
		return p2, nil, err
	}

	innerData := EncodeTLVList(currencyTypeTLV, lessThanTLV, greaterThanTLV)
	cmdData, err := NewTLV(TagPhononFilter, innerData)
	if err != nil {
		return p2, nil, err
	}

	return p2, cmdData.Encode(), nil
}

func parseListPhononsResponse(resp []byte) ([]model.Phonon, error) {
	phononCollection, err := ParseTLVPacket(resp, TagPhononCollection)
	if err != nil {
		return nil, err
	}
	//No phonons in list, the only tag will be the overall collection
	if len(phononCollection) <= 1 {
		return nil, nil
	}
	phonons := make([]model.Phonon, 0)
	phononDescriptions, err := phononCollection.FindTags(TagPhononDescriptor)
	if err != nil {
		return nil, err
	}

	for _, description := range phononDescriptions {
		descriptionTLV, err := ParseTLVPacket(description)
		if err != nil {
			return phonons, err
		}
		keyIndexBytes, err := descriptionTLV.FindTag(TagKeyIndex)
		if err != nil {
			return phonons, err
		}
		currencyTypeBytes, err := descriptionTLV.FindTag(TagCurrencyType)
		if err != nil {
			return phonons, err
		}
		currencyType := binary.BigEndian.Uint16(currencyTypeBytes)

		valueBytes, err := descriptionTLV.FindTag(TagPhononValue)
		if err != nil {
			return phonons, err
		}

		var value float32
		err = binary.Read(bytes.NewReader(valueBytes), binary.BigEndian, &value)
		if err != nil {
			return phonons, err
		}
		phonon := model.Phonon{
			KeyIndex:     int(binary.BigEndian.Uint16(keyIndexBytes)),
			CurrencyType: model.CurrencyType(currencyType),
			Value:        value,
		}
		phonons = append(phonons, phonon)
	}
	return phonons, nil
}

//Generally checks status, including extended responses
func checkStatusWord(status uint16) (continues bool, err error) {
	if status == 0x9000 {
		return false, nil
	}
	if status > 0x9000 {
		return true, nil
	}
	//TODO: Add error conditions
	return false, ErrUnknown
}

func (cs *PhononCommandSet) GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error) {
	data, err := NewTLV(TagKeyIndex, util.Uint16ToBytes(keyIndex))
	if err != nil {
		return nil, err
	}
	cmd := NewCommandGetPhononPubKey(data.Encode())
	resp, err := cs.c.Send(cmd)
	if err != nil {
		return nil, err
	}

	pubKey, err := parseGetPhononPubKeyResponse(resp.Data)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func parseGetPhononPubKeyResponse(resp []byte) (pubKey *ecdsa.PublicKey, err error) {
	collection, err := ParseTLVPacket(resp)
	if err != nil {
		return nil, err
	}
	rawPubKey, err := collection.FindTag(TagPhononPubKey)
	if err != nil {
		return nil, err
	}

	pubKey, err = util.ParseECDSAPubKey(rawPubKey)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (cs *PhononCommandSet) DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error) {
	data, err := NewTLV(TagKeyIndex, util.Uint16ToBytes(keyIndex))
	if err != nil {
		return nil, err
	}
	cmd := NewCommandDestroyPhonon(data.Encode())
	resp, err := cs.c.Send(cmd)
	if err != nil {
		return nil, err
	}
	err = cs.checkOK(resp, err)
	if err != nil {
		return nil, err
	}
	//parse private key from response

	return nil, nil
}

func (cs *PhononCommandSet) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets [][]byte, err error) {
	//Save this for extended requests
	// tlvLength := 2
	// bytesPerKeyIndex := 2
	// apduHeaderLength := 4
	// maxPhononsPerRequest := (maxAPDULength - apduHeaderLength - tlvLength) / bytesPerKeyIndex
	// numPhonons := len(keyIndices)
	// remainingKeyIndices := make([]uint16, 0)
	// if numPhonons > maxPhononsPerRequest {
	// 	remainingKeyIndices = keyIndices[maxPhononsPerRequest:]
	// 	keyIndices = keyIndices[:maxPhononsPerRequest]
	// }

	//TODO: protect the caller from passing too many keyIndices for an APDU
	cmd := NewCommandSendPhonons(keyIndices, extendedRequest)
	resp, err := cs.c.Send(cmd)
	if err != nil {
		log.Error("error in send phonons command: ", err)
		return nil, err
	}

	//TODO: replace with specific status check
	//TODO: parse response for extended status code
	continues, err := checkStatusWord(resp.Sw)
	if err != nil {
		return nil, err
	}

	transferPhononPackets = append(transferPhononPackets, resp.Data)

	//Recursively call the extended list and append the result packets to
	var remainingPhononPackets [][]byte
	if continues {
		remainingPhononPackets, err = cs.SendPhonons(nil, true)
		if err != nil {
			return nil, err
		}
	}
	for _, packet := range remainingPhononPackets {
		transferPhononPackets = append(transferPhononPackets, packet)
	}

	//Maybe save this for extended request form
	// //Redo this to receive multiple responses, not to send multiple requests
	// //Recursively call SendPhonons until all extended requests and responses are receivedy
	// if len(remainingKeyIndices) > 0 {
	// 	extendedPhononPackets, err := cs.SendPhonons(remainingKeyIndices, false)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	transferPhononPackets = append(transferPhononPackets, extendedPhononPackets...)
	// }
	return transferPhononPackets, nil
}
