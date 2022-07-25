package card

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/GridPlus/keycard-go"
	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/gridplus"
	"github.com/GridPlus/keycard-go/types"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/config"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/tlv"
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

var apduLogFile *os.File
var apduLogger *log.Logger

type PhononCommandSet struct {
	c               types.Channel
	sc              *SecureChannel
	ApplicationInfo *types.ApplicationInfo
	PairingInfo     *types.PairingInfo
	PhononCACert    []byte
}

func NewPhononCommandSet(c types.Channel) *PhononCommandSet {
	var err error
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("could not configure PhononCommandSet")
	}
	var level = conf.LogLevel

	if level == log.DebugLevel {
		//Create an apdu.log file in the current working directory
		dir, err := os.Getwd()
		if err != nil {
			log.Error("could not fetch working directory")
		} else {
			apduLogFile, err = os.Create(filepath.FromSlash(dir + "/apdu.log"))
			if err != nil {
				log.Error("failed to create apdu.log", err)
			} else {
				log.Info("created apdu.log")
			}
		}

	}
	apduLogger = &log.Logger{
		Out:       apduLogFile,
		Formatter: &APDUDebugFormatter{},
		Level:     log.DebugLevel,
	}

	return &PhononCommandSet{
		c:               c,
		sc:              NewSecureChannel(c),
		ApplicationInfo: &types.ApplicationInfo{},
		PhononCACert:    conf.AppletCACert,
	}
}

func (cs PhononCommandSet) Send(cmd *Command) (*apdu.Response, error) {
	//Log commands to apdu log
	//Log APDUs in debugger format to file
	apduLogger.Debugf("#INS % X\n", cmd.ApduCmd.Ins)
	outputAPDU, _ := cmd.ApduCmd.Serialize()
	apduLogger.Debugf("/send %X\n", outputAPDU)

	resp, err := cs.c.Send(cmd.ApduCmd)
	if err != nil {
		return resp, err
	}
	err = cmd.HumanReadableErr(resp)
	return resp, err

}

//Selects the phonon applet for further usage
func (cs *PhononCommandSet) Select() (instanceUID []byte, cardPubKey *ecdsa.PublicKey, cardInitialized bool, err error) {
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
	instanceUID, cardPubKey, cardInitialized, err = parseSelectResponse(resp.Data)
	if err != nil {
		log.Error("error parsing select response. err: ", err)
		return nil, nil, false, err
	}

	//Generate secure channel secrets using card's public key
	secretsErr := cs.sc.GenerateSecret(ethcrypto.FromECDSAPub(cardPubKey))
	if secretsErr != nil {
		log.Error("could not generate secure channel secrets. err: ", secretsErr)
		return nil, nil, true, secretsErr
	}
	log.Debugf("Pairing generated key: % X\n", cs.sc.RawPublicKey())

	return instanceUID, cardPubKey, cardInitialized, nil
}

func (cs *PhononCommandSet) Pair() (*cert.CardCertificate, error) {
	log.Debug("sending PAIR command")
	//Generate random salt and keypair
	clientSalt := make([]byte, 32)
	rand.Read(clientSalt)

	pairingPrivKey, err := ethcrypto.GenerateKey()
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

	cardCertPubKey, err := util.ParseECCPubKey(cardCert.PubKey)
	if err != nil {
		return &cert.CardCertificate{}, err
	}
	//Validate card's certificate has valid GridPlus signature
	err = cert.ValidateCardCertificate(cardCert, cs.PhononCACert)
	if err != nil {
		log.Error("unable to verify card certificate signature")
		return &cert.CardCertificate{}, err
	}
	log.Debug("certificate signature valid")

	pubKeyValid := gridplus.ValidateECCPubKey(cardCertPubKey)
	log.Debug("certificate public key valid: ", pubKeyValid)
	if !pubKeyValid {
		log.Error("card pubkey invalid")
		return &cert.CardCertificate{}, errors.New("certificate pubkey invalid")
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
		switch pairingStep {
		case 1:
			err = errors.New("unable to generate secret")
		case 2:
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

/*OpenSecureChannel is a convenience function to perform all of the necessary options to open a card
to terminal secure channel in sequence. Runs SELECT, PAIR, OPEN_SECURE_CHANNEL*/
func (cs *PhononCommandSet) OpenSecureConnection() error {
	_, _, initialized, err := cs.Select()
	if err != nil {
		log.Error("could not select phonon applet: ", err)
		return err
	}
	if !initialized {
		return ErrCardUninitialized
	}
	_, err = cs.Pair()
	if err != nil {
		log.Error("could not pair: ", err)
		return err
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		log.Error("could not open secure channel: ", err)
		return err
	}
	return nil
}

func (cs *PhononCommandSet) OpenBestConnection() (initialized bool, err error) {
	_, _, initialized, err = cs.Select()
	if !initialized {
		return false, err
	}
	_, err = cs.Pair()
	if err != nil {
		return false, err
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		return false, err
	}
	return initialized, nil
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
	log.Debug("sending IDENTIFY_CARD command")
	cmd := NewCommandIdentifyCard(nonce)
	resp, err := cs.Send(cmd)
	if err != nil {
		log.Error("could not send identify card command", err)
		return nil, nil, err
	}

	cardPubKey, cardSig, err = ParseIdentifyCardResponse(resp.Data)
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

func (cs *PhononCommandSet) CreatePhonon(curveType model.CurveType) (keyIndex uint16, pubKey model.PhononPubKey, err error) {
	log.Debug("sending CREATE_PHONON command")

	cmd := NewCommandCreatePhonon(byte(curveType))
	resp, err := cs.sc.Send(cmd) //temp normal channel for testing
	if err != nil {
		log.Error("create phonon command failed: ", err)
		return 0, nil, err
	}
	if err = checkPhononTableErrors(resp.Sw); err != nil {
		return 0, nil, err
	}
	keyIndex, pubKeyBytes, err := parseCreatePhononResponse(resp.Data)
	if err != nil {
		return 0, nil, err
	}

	pubKey, err = model.NewPhononPubKey(pubKeyBytes, curveType)
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

func (cs *PhononCommandSet) SetDescriptor(p *model.Phonon) error {
	log.Debug("sending SET_DESCRIPTOR command")
	data, err := encodeSetDescriptorData(p)
	if err != nil {
		return err
	}

	cmd := NewCommandSetDescriptor(data)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		log.Error("set descriptor command failed: ", err)
		return err
	}

	return cs.checkOK(resp, err)
}

//ListPhonons takes a currency type and range bounds and returns a listing of the phonons currently stored on the card
//Set lessThanValue or greaterThanValue to 0 to ignore the parameter. Returned phonons omit the public key to reduce data transmission
//After processing, the list client should send GET_PHONON_PUB_KEY to retrieve the corresponding pubkeys if necessary.
func (cs *PhononCommandSet) ListPhonons(currencyType model.CurrencyType, lessThanValue uint64, greaterThanValue uint64, continuation bool) ([]*model.Phonon, error) {
	log.Debug("sending LIST_PHONONS command")
	p2, cmdData, err := encodeListPhononsData(currencyType, lessThanValue, greaterThanValue)
	if err != nil {
		return nil, err
	}
	log.Debug("List phonons command data: ")
	log.Debugf("% X", cmdData)
	log.Debugf("p2: % X", p2)
	var p1 byte
	if continuation {
		p1 = 0x01
	} else {
		p1 = 0x00
	}
	cmd := NewCommandListPhonons(p1, p2, cmdData)
	resp, err := cs.sc.Send(cmd)
	if err != nil && err != ErrDefault {
		log.Error("error in sending listPhonons. err: ", err)
		return nil, err
	}

	err = checkPhononTableErrors(resp.Sw)
	if err != nil {
		log.Error("phonon table error detected: ", err)
		return nil, err
	}

	continuation, err = checkContinuation(resp.Sw)
	if err != nil {
		log.Error("error detected while checking for list continuation. err: ", err)
		return nil, err
	}

	phonons, err := parseListPhononsResponse(resp.Data)
	if err != nil {
		log.Error("could not parse list phonons response: ", err)
		return nil, err
	}
	if continuation {
		extendedPhonons, err := cs.ListPhonons(currencyType, lessThanValue, greaterThanValue, continuation)
		if err != nil {
			log.Error("could not read extended phonons list: ", err)
			return nil, err
		}
		phonons = append(phonons, extendedPhonons...)
	}
	return phonons, nil
}

//Generally checks status, including extended responses
func checkContinuation(status uint16) (continues bool, err error) {
	if status == 0x9000 {
		return false, nil
	}
	if status > 0x9000 && status < 0x9100 {
		return true, nil
	}
	return false, ErrUnknown
}

func (cs *PhononCommandSet) GetPhononPubKey(keyIndex uint16, crv model.CurveType) (pubKey model.PhononPubKey, err error) {
	log.Debug("sending GET_PHONON_PUB_KEY command")
	data, err := tlv.NewTLV(TagKeyIndex, util.Uint16ToBytes(keyIndex))
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

	rawPubKey, err := parseGetPhononPubKeyResponse(resp.Data)
	if err != nil {
		return nil, err
	}
	pubKey, err = model.NewPhononPubKey(rawPubKey, crv)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}

func (cs *PhononCommandSet) DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error) {
	log.Debug("sending DESTROY_PHONON command")
	data, err := tlv.NewTLV(TagKeyIndex, util.Uint16ToBytes(keyIndex))
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

func (cs *PhononCommandSet) SendPhonons(keyIndices []uint16, extendedRequest bool) (transferPhononPackets []byte, err error) {
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

	transferPhononPackets = append(transferPhononPackets, resp.Data...)

	//Recursively call the extended list and append the result packets to
	var remainingPhononPackets []byte
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
func (cs *PhononCommandSet) InitCardPairing(receiverCert cert.CardCertificate) (initPairingData []byte, err error) {
	log.Debug("sending INIT_CARD_PAIRING command")
	certTLV, err := tlv.NewTLV(TagCardCertificate, receiverCert.Serialize())
	if err != nil {
		return nil, err
	}
	cmd := NewCommandInitCardPairing(certTLV.Encode())
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

	return resp.Data, nil
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
	nonce := make([]byte, 32)
	n, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return fmt.Errorf("unable to retrieve random challenge for card: %s", err.Error())
	}
	if n != 32 {
		return fmt.Errorf("unable to read 32 bytes for challenge to card")
	}

	// Send Challenge to card
	cardPubKey, sig, err := cs.IdentifyCard(nonce)
	if err != nil {
		return fmt.Errorf("unable to identify card %s", err.Error())
	}

	sigValid := ecdsa.Verify(cardPubKey, nonce, sig.R, sig.S)
	if !sigValid {
		return errors.New("invalid signature over nonce")
	}

	signedCert, err := cert.CreateCardCertificate(cardPubKey, signKeyFunc)
	if err != nil {
		return err
	}

	log.Debug("sending INSTALL_CERTIFICATE command")
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

func (cs *PhononCommandSet) GetFriendlyName() (string, error) {
	log.Debug("sending GET_FRIENDLY_NAME command")
	cmd := NewCommandGetFriendlyName()
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return "", err
	}
	name := string(resp.Data)
	return name, nil
}

func (cs *PhononCommandSet) SetFriendlyName(name string) error {
	log.Debug("sending SET_FRIENDLY_NAME command")
	cmd := NewCommandSetFriendlyName(name)
	_, err := cs.sc.Send(cmd)
	return err
}

func (cs *PhononCommandSet) GetAvailableMemory() (persistentMem int, onResetMem int, onDeselectMem int, err error) {
	log.Debug("sending GET_AVAILABLE_MEMORY command")
	cmd := NewCommandGetAvailableMemory()
	data, err := cs.sc.Send(cmd)
	if err != nil {
		return 0, 0, 0, err
	}

	persistentMem, onResetMem, onDeselectMem, err = parseGetAvailableMemoryResponse(data.Data)
	if err != nil {
		return 0, 0, 0, err
	}
	return persistentMem, onResetMem, onDeselectMem, nil
}

func (cs *PhononCommandSet) MineNativePhonon(difficulty uint8) (keyIndex uint16, hash []byte, err error) {
	log.Debug("sending MINE_NATIVE_PHONON command")
	cmd := NewCommandMineNativePhonon(difficulty)
	resp, err := cs.sc.Send(cmd)
	if err != nil {
		return 0, nil, err
	}
	keyIndex, hash, err = parseMineNativePhononResponse(resp.Data)
	if err != nil {
		return 0, nil, err
	}

	return keyIndex, hash, nil
}
