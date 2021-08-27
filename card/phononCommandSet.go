package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/GridPlus/keycard-go"
	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/keycard-go/types"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

const (
	StatusSuccess         = 0x9000
	StatusPhononTableFull = 0x6A84
	StatusKeyIndexInvalid = 0x6983
	StatusOutOfMemory     = 0x6F00
	StatusPINNotEntered   = 0x6985
)

var (
	ErrCardUninitialized = errors.New("card uninitialized")
	ErrPhononTableFull   = errors.New("phonon table full")
	ErrKeyIndexInvalid   = errors.New("key index out of valid range")
	ErrOutOfMemory       = errors.New("card out of memory")
	ErrPINNotEntered     = errors.New("valid PIN required")
	ErrUnknown           = errors.New("unknown error")
)

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

func (cs PhononCommandSet) Send(cmd *Command) (*apdu.Response, error) {
	resp, err := cs.c.Send(cmd.ApduCmd)
	if err != nil {
		return resp, err
	}
	err = cmd.HumanReadableErr(resp)
	return resp, err

}

//TODO: determine if I should return these values or have the secure channel handle it internally
//Selects the phonon applet for further usage
func (cs *PhononCommandSet) Select() (instanceUID []byte, cardPubKey []byte, cardInitialized bool, err error) {
	cmd := NewCommandSelectPhononApplet()
	cmd.ApduCmd.SetLe(0)

	log.Debug("sending SELECT apdu")
	resp, err := cs.Send(cmd)
	if err != nil {
		log.Error("could not send select command. err: ", err)
		return nil, nil, false, err
	}

	err = cs.checkOK(resp, err)
	if err != nil {
		return nil, nil, false, err
	}
	instanceUID, cardPubKey, err = parseSelectResponse(resp.Data)
	if err != nil && err != ErrCardUninitialized {
		log.Error("error parsing select response. err: ", err)
		return nil, nil, false, err
	}

	//TODO: Use random version GenerateSecret in production
	//Generate secure channel secrets using card's public key
	secretsErr := cs.sc.GenerateSecret(cardPubKey)
	if secretsErr != nil {
		log.Error("could not generate secure channel secrets. err: ", secretsErr)
		return nil, nil, true, secretsErr
	}
	log.Debugf("Pairing generated key: % X\n", cs.sc.RawPublicKey())
	//return ErrCardUninitialized if ParseSelectResponse returns that error code
	if err == ErrCardUninitialized {
		return instanceUID, cardPubKey, false, nil
	}
	return instanceUID, cardPubKey, true, nil
}

func (cs *PhononCommandSet) Pair() error {
	log.Debug("sending PAIR command")
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
	cmd := NewCommandPairStep1(clientSalt, &pairingPubKey)

	resp, err := cs.Send(cmd)
	if err != nil {
		log.Error("unable to send Pair Step 1 command. err: ", err)
		return err
	}
	err = checkPairingErrors(1, resp.Sw)
	if err != nil {
		return err
	}
	pairStep1Resp, err := gridplus.ParsePairStep1Response(resp.Data)
	if err != nil {
		log.Error("could not parse pair step 1 response. err: ", err)
		return err
	}

	//Validate card's certificate has valid GridPlus signature
	certValid := gridplus.ValidateCardCertificate(pairStep1Resp.SafecardCert, gridplus.SafecardDevCAPubKey)
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
		return err
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
	cmd = NewCommandPairStep2(cryptogram)
	resp, err = cs.Send(cmd)
	if err != nil {
		log.Error("error sending pair step 2 command. err: ", err)
		return err
	}
	err = checkPairingErrors(2, resp.Sw)
	if err != nil {
		return err
	}
	pairStep2Resp, err := gridplus.ParsePairStep2Response(resp.Data)
	if err != nil {
		log.Error("could not parse pair step 2 response. err: ", err)
		return err
	}
	log.Debugf("pairStep2Resp: % X", pairStep2Resp)

	//Derive Pairing Key
	pairingKey := sha256.Sum256(append(pairStep2Resp.Salt, secretHash...))
	log.Debugf("derived pairing key: % X", pairingKey)

	//Store pairing info for use in OpenSecureChannel
	cs.setPairingInfo(pairingKey[0:], pairStep2Resp.PairingIdx)

	log.Debug("pairing succeeded")
	return nil
}

//checkPairingErrors takes a pairing step, either 1 or 2, and the SW value of the response to return appropriate error messages
//Errors and error codes are defined internally to this function as they are specific to PAIR and do not apply to other commands
func checkPairingErrors(pairingStep int, status uint16) (err error) {
	if pairingStep != 1 && pairingStep != 2 {
		return errors.New("pairing step must be set to 1 or 2 to check pairing errors")
	}
	switch status {
	case 0x9000:
		err = nil
	case 0x6A80:
		err = errors.New("invalid pairing data format")
	case 0x6882:
		err = errors.New("certificate not loaded")
	case 0x6982:
		if pairingStep == 1 {
			err = errors.New("unable to generate secret")
		} else if pairingStep == 2 {
			err = errors.New("client cryptogram verification failed")
		}
	case 0x6A84:
		err = errors.New("all available pairing slots taken")
	case 0x6A86:
		err = errors.New("p1 invalid or first pairing phase was not completed")
	case 0x6985:
		err = errors.New("secure channel is already open")
	case 0x6D00:
		err = errors.New("pin has not been set")
	}

	return err
}

func (cs *PhononCommandSet) setPairingInfo(key []byte, index int) {
	cs.PairingInfo = &types.PairingInfo{
		Key:   key,
		Index: index,
	}
}

func (cs *PhononCommandSet) Unpair(index uint8) error {
	log.Debug("sending UNPAIR command")
	cmd := NewCommandUnpair(index)
	resp, err := cs.sc.Send(cmd)
	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) OpenSecureChannel() error {
	log.Debug("sending OPEN_SECURE_CHANNEL command")
	if cs.ApplicationInfo == nil {
		return errors.New("cannot open secure channel without setting PairingInfo")
	}

	cmd := NewCommandOpenSecureChannel(uint8(cs.PairingInfo.Index), cs.sc.RawPublicKey())
	resp, err := cs.Send(cmd)
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
	log.Debug("sending MUTUAL_AUTH command")
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return err
	}

	cmd := NewCommandMutualAuthenticate(data)

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
	init := NewCommandInit(data)
	resp, err := cs.Send(init)

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

func (cs *PhononCommandSet) IdentifyCard(nonce []byte) (cardPubKey *ecdsa.PublicKey, cardSig *util.ECDSASignature, err error) {
	cmd := NewCommandIdentifyCard(nonce)
	resp, err := cs.Send(cmd)
	if err != nil {
		log.Error("could not send identify card command", err)
		return nil, nil, err
	}
	log.Debug("identify card resp:\n", hex.Dump(resp.Data))

	cardPubKey, cardSig, err = parseIdentifyCardResponse(resp.Data)
	if err != nil {
		log.Error("could not parse identify card response: ", err)
		return nil, nil, err
	}

	valid := ecdsa.Verify(cardPubKey, nonce, cardSig.R, cardSig.S)
	if !valid {
		return cardPubKey, cardSig, errors.New("card signature over challenge salt is invalid")
	}

	return cardPubKey, cardSig, nil
}

func (cs *PhononCommandSet) VerifyPIN(pin string) error {
	log.Debug("sending VERIFY_PIN command")
	cmd := NewCommandVerifyPIN(pin)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		log.Error("could not send VERIFY_PIN command", err)
		return err
	}
	triesRemaining, err := checkVerifyPINErrors(resp.Sw)
	if err != nil {
		log.Error("error verifying pin: ", err)
		log.Error("triesRemaining: ", triesRemaining)
		return err
	}
	return nil
}

func checkVerifyPINErrors(status uint16) (triesRemaining int, err error) {
	if status >= 0x63C0 && status < 0x63D0 {
		triesRemaining = int(status - 0x63C0)
		return triesRemaining, errors.New("incorrect pin")
	}
	return 0, nil
}

func (cs *PhononCommandSet) ChangePIN(pin string) error {
	log.Debug("sending CHANGE_PIN command")
	cmd := NewCommandChangePIN(pin)
	resp, err := cs.sc.Send(cmd)

	return cs.checkOK(resp, err)
}

func (cs *PhononCommandSet) CreatePhonon() (keyIndex uint16, pubKey *ecdsa.PublicKey, err error) {
	log.Debug("sending CREATE_PHONON command")
	cmd := NewCommandCreatePhonon()
	resp, err := cs.sc.Send(cmd) //temp normal channel for testing
	if err != nil {
		log.Error("create phonon command failed: ", err)
		return 0, nil, err
	}
	if err = checkPhononTableErrors(resp.Sw); err != nil {
		return 0, nil, err
	}
	keyIndex, pubKey, err = parseCreatePhononResponse(resp.Data)
	if err != nil {
		return 0, nil, err
	}
	return keyIndex, pubKey, nil
}

//Common error code checks for commands that deal with the phonon table
//CREATE_PHONON, LIST_PHONONS, DESTROY_PHONON, etc.
func checkPhononTableErrors(sw uint16) error {
	switch sw {
	case StatusPhononTableFull:
		return ErrPhononTableFull
	case StatusOutOfMemory:
		return ErrOutOfMemory
	case StatusKeyIndexInvalid:
		return ErrKeyIndexInvalid
	case StatusPINNotEntered:
		return ErrPINNotEntered
	case StatusSuccess:
		return nil
	}
	return nil
}

func (cs *PhononCommandSet) SetDescriptor(keyIndex uint16, currencyType model.CurrencyType, value float32) error {
	log.Debug("sending SET_DESCRIPTOR command")
	data, err := encodeSetDescriptorData(keyIndex, currencyType, value)
	if err != nil {
		return err
	}

	cmd := NewCommandSetDescriptor(data)
	resp, err := cs.sc.Send(cmd) //temp normal channel for testing
	if err != nil {
		log.Error("set descriptor command failed: ", err)
		return err
	}

	return cs.checkOK(resp, err)
}

//ListPhonons takes a currency type and range bounds and returns a listing of the phonons currently stored on the card
//Set lessThanValue or greaterThanValue to 0 to ignore the parameter. Returned phonons omit the public key to reduce data transmission
//After processing, the list client should send GET_PHONON_PUB_KEY to retrieve the corresponding pubkeys if necessary.
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
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}

	err = checkPhononTableErrors(resp.Sw)
	if err != nil {
		return nil, err
	}

	continues, err := checkContinuation(resp.Sw)
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
	log.Debug("sending LIST_PHONONS extended request")
	cmd := NewCommandListPhonons(0x01, 0x00, nil)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}
	continues, err := checkContinuation(resp.Sw)
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

//TODO: replace this check in send and receive and other functions with specific error checks
//Generally checks status, including extended responses
func checkContinuation(status uint16) (continues bool, err error) {
	if status == 0x9000 {
		return false, nil
	}
	if status > 0x9000 {
		return true, nil
	}
	return false, ErrUnknown
}

func (cs *PhononCommandSet) GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error) {
	log.Debug("sending GET_PHONON_PUB_KEY command")
	data, err := NewTLV(TagKeyIndex, util.Uint16ToBytes(keyIndex))
	if err != nil {
		return nil, err
	}
	cmd := NewCommandGetPhononPubKey(data.Encode())
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}

	err = checkPhononTableErrors(resp.Sw)
	if err != nil {
		return nil, err
	}

	pubKey, err := parseGetPhononPubKeyResponse(resp.Data)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (cs *PhononCommandSet) DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error) {
	log.Debug("sending DESTROY_PHONON command")
	data, err := NewTLV(TagKeyIndex, util.Uint16ToBytes(keyIndex))
	if err != nil {
		return nil, err
	}
	cmd := NewCommandDestroyPhonon(data.Encode())
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}
	err = cs.checkOK(resp, err)
	if err != nil {
		return nil, err
	}
	privKey, err = parseDestroyPhononResponse(resp.Data)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func (cs *PhononCommandSet) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets [][]byte, err error) {
	log.Debug("sending SEND_PHONONS command")
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
	data, p2Length := encodeSendPhononsData(keyIndices)
	cmd := NewCommandSendPhonons(data, p2Length, extendedRequest)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		log.Error("error in send phonons command: ", err)
		return nil, err
	}

	continues, err := checkContinuation(resp.Sw)
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
	transferPhononPackets = append(transferPhononPackets, remainingPhononPackets...)

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

func (cs *PhononCommandSet) ReceivePhonons(phononTransfer []byte) error {
	log.Debug("sending RECV_PHONONS command")

	cmd := NewCommandReceivePhonons(phononTransfer)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return err
	}
	return nil
}

//Implemented with support for single
func (cs *PhononCommandSet) SetReceiveList(phononPubKeys []*ecdsa.PublicKey) error {
	log.Debug("sending SET_RECV_LIST command")
	data, err := encodeSetReceiveListData(phononPubKeys)
	if err != nil {
		return err
	}
	cmd := NewCommandSetReceiveList(data)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return err
	}
	return nil
}

func (cs *PhononCommandSet) TransactionAck(keyIndices []uint16) error {
	log.Debug("sending TRANSACTION_ACK command")

	data := encodeKeyIndexList(keyIndices)

	cmd := NewCommandTransactionAck(data)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return err
	}
	return nil
}

//InitCardPairing tells a card to initialized a pairing with another phonon card
//Data is passed transparently from card to card since no client processing is necessary
func (cs *PhononCommandSet) InitCardPairing() (initPairingData []byte, err error) {
	log.Debug("sending INIT_CARD_PAIRING command")
	cmd := NewCommandInitCardPairing()
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

//CardPair takes the response from initCardPairing and passes it to the counterparty card
//for the next step of pairing
func (cs *PhononCommandSet) CardPair(initPairingData []byte) (cardPairData []byte, err error) {
	log.Debug("sending CARD_PAIR command")
	cmd := NewCommandCardPair(initPairingData)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (cs *PhononCommandSet) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	log.Debug("sending CARD_PAIR_2 command")
	cmd := NewCommandCardPair2(cardPairData)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return nil, err
	}

	return cardPair2Data, nil
}

func (cs *PhononCommandSet) FinalizeCardPair(cardPair2Data []byte) (err error) {
	log.Debug("sending FINALIZE_CARD_PAIR command")
	cmd := NewCommandFinalizeCardPair(cardPair2Data)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return err
	}
	_, err = checkContinuation(resp.Sw)
	if err != nil {
		return err
	}

	return nil
}

func (cs *PhononCommandSet) InstallCertificate(signKeyFunc func([]byte) ([]byte, error)) (err error) {
	log.Debug("sending INSTALL_CERTIFICATE command")
	nonce := make([]byte, 32)
	n, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return fmt.Errorf("unable to retrieve random challenge for card: %s", err.Error())
	}
	if n != 32 {
		return fmt.Errorf("unable to read 32 bytes for challenge to card")
	}

	// Send Challenge to card
	cardPubKey, _, err := cs.IdentifyCard(nonce)
	if err != nil {
		return fmt.Errorf("unable to identify card %s", err.Error())
	}

	signedCert, err := createCardCertificate(cardPubKey, signKeyFunc)
	if err != nil {
		return err
	}

	cmd := NewCommandInstallCert(signedCert)
	resp, err := cs.Send(cmd)
	if err != nil {
		return err
	}
	err = checkInstallCertError(resp.Sw)
	if err != nil {
		return err
	}

	return nil

}

func checkInstallCertError(status uint16) error {
	switch status {
	case 0x9000:
		return nil
	case 0x6986:
		return errors.New("certificate already loaded")
	case 0x6984:
		return errors.New("data invalid")
	default:
		return errors.New("unknown error")
	}
}

func (cs *PhononCommandSet) GenerateInvoice() (invoiceData []byte, err error) {
	log.Debug("sending GENERATE_INVOICE command")
	cmd := NewCommandGenerateInvoice()
	res, err := cs.sc.Send(cmd)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (cs *PhononCommandSet) ReceiveInvoice(invoiceData []byte) (err error) {
	log.Debug("sending RECEIVE_INVOICE command")
	cmd := NewCommandReceiveInvoice()
	_, err = cs.sc.Send(cmd)
	if err != nil {
		return err
	}

	return nil
}
