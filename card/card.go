package card

import (
	"errors"
	"fmt"

	"github.com/GridPlus/keycard-go"
	"github.com/GridPlus/keycard-go/io"
	"github.com/GridPlus/phonon-client/util"
	"github.com/ebfe/scard"
	log "github.com/sirupsen/logrus"
)

//Copied from safecard.c AID_SAFECARD_V1
var (
	SafecardAID                                   = []byte{0xA0, 0x00, 0x00, 0x08, 0x20, 0x00, 0x01, 0x01}
	StatusWalletAID                               = []byte{0xA0, 0x00, 0x00, 0x08, 0x04, 0x00, 0x01, 0x01, 0x01}
	SAFECARD_APDU_CLA_ENCRYPTED_PROPRIETARY uint8 = 0x80
	SAFECARD_APDU_INS_PAIR                  uint8 = 0x12
	SAFECARD_APDU_INS_OPEN_SECURE_CHANNEL   uint8 = 0x10
	PAIR_STEP1                              uint8 = 0x00
	PAIR_STEP2                              uint8 = 0x01
	TLV_TYPE_CUSTOM                         uint8 = 0x80

	// //PairStep1 Constants
	// sigLength               int = 74
	// certLength              int = 147 //8 TLV Header + 65 PubKey + 74 Sig
	// saltSize                int = 32
	// correctPairStep1RespLen int = saltSize + sigLength + certLength
)

type Safecard struct {
	*keycard.CommandSet
	// c           *io.NormalChannel
	// sc          *keycard.SecureChannel
	// ctx         *scard.Context
	// card        *scard.Card
	// instanceUID []byte
	// pubKey      []byte
	// privKey     *secp256k1.PrivateKey
	// pairingKey  []byte //Key derived after successful pairing
}

// type Safecard keycard.CommandSet

func Connect() (Safecard, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		fmt.Println(err)
		return Safecard{}, err
	}
	// defer func() {
	// 	if err := ctx.Release(); err != nil {
	// 		log.Error("error releasing context", err)
	// 	}
	// }()

	readers, err := ctx.ListReaders()
	if err != nil {
		fmt.Println(err)
		return Safecard{}, err
	}

	for i, reader := range readers {
		fmt.Printf("[%d] %s\n", i, reader)
	}

	if len(readers) > 0 {
		card, err := ctx.Connect(readers[0], scard.ShareShared, scard.ProtocolAny)
		if err != nil {
			log.Error(err)
		}
		// defer card.Disconnect(scard.ResetCard)

		fmt.Println("Card status:")
		status, err := card.Status()
		if err != nil {
			log.Error(err)
		}

		fmt.Printf("\treader: %s\n\tstate: %x\n\tactive protocol: %x\n\tatr: % x\n",
			status.Reader, status.State, status.ActiveProtocol, status.Atr)

		// c.c = io.NewNormalChannel(card)
		// //Set card context
		// c.ctx = ctx
		// c.card = card
		return Safecard{
			keycard.NewCommandSet(io.NewNormalChannel(card)),
		}, nil
	}
	return Safecard{}, errors.New("no card reader found")
}

// func (c Safecard) ExportSeed() ([]byte, error) {

// 	// var cmd = []byte{0x00, 0xa4, 0x00, 0x0c, 0x02, 0x3f, 0x00} // SELECT MF

// 	var cmd = []byte{0x80, 0xC3, 0x00, 0x00}
// 	fmt.Println("Transmit:")
// 	fmt.Printf("\tc-apdu: % x\n", cmd)
// 	rsp, err := c.card.Transmit(cmd)
// 	if err != nil {
// 		log.Error("error transmitting apdu", err)
// 		return nil, err
// 	}
// 	fmt.Printf("\tr-apdu: % x\n", rsp)
// 	return rsp, nil
// }

//Mock Card function
func (c Safecard) CreatePhonons(n int) (pubKeys [][]byte, err error) {
	phononPubKeys := make([][]byte, 0)
	for i := 0; i < n; i++ {
		phononPubKeys = append(phononPubKeys, util.RandomKey(32))
	}
	return phononPubKeys, nil
}

// //Select loads the GridPlus safecard applet
// func (c Safecard) Select() error {
// 	cmd := globalplatform.NewCommandSelect(SafecardAID)
// 	cmd.SetLe(0)
// 	resp, err := c.Send(cmd)
// 	if err != nil {
// 		log.Error("could not send select command. err: ", err)
// 		return err
// 	}
// 	// apdu, err := cmd.Serialize()
// 	// if err != nil {
// 	// 	log.Error("could not serialize apdu. err: ", err)
// 	// }
// 	// resp, err := c.card.Transmit(apdu)
// 	// if err != nil {
// 	// 	log.Error("error sending select. err: ", err)
// 	// 	return err
// 	// }

// 	instanceUID, pubKey, err := parseSelectResponse(resp.Data)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infof("instanceUID: % X\npubKey: % X", instanceUID, pubKey)
// 	c.instanceUID = instanceUID
// 	c.pubKey = pubKey

// 	log.Debug("select response: % X", resp)
// 	return nil
// }

// //Pair creates a secure pairing between the terminal and the card
// func (c Safecard) Pair() error {
// 	//Generate random client salt
// 	clientSalt := make([]byte, 32)
// 	rand.Read(clientSalt)

// 	clientPrivKey, err := secp256k1.GeneratePrivateKey()
// 	if err != nil {
// 		log.Error("could not generate ECC private key")
// 	}

// 	pairStep1Resp, err := c.SendPairStep1(clientSalt, &clientPrivKey.PublicKey)

// 	// log.Infof("cardSalt: % X", pairStep1cardSalt)
// 	// log.Infof("safecardCert: % X", safecardCert)
// 	// log.Infof("cardSig: len(%v)\n % X", len(cardSig), cardSig)
// 	certValid := validateCardCertificate(pairStep1Resp.safecardCert)
// 	log.Info("certificate signature valid: ", certValid)
// 	if !certValid {
// 		log.Error("unable to verify card certificate.")
// 		return err
// 	}

// 	//Parse ECDSA pubkey object from cardPubKey bytes
// 	//Offset start of pubkey by 3
// 	//2 byte TLV header + DER type byte
// 	cardPubKey := &ecdsa.PublicKey{
// 		Curve: secp256k1.S256(),
// 		X:     new(big.Int).SetBytes(pairStep1Resp.safecardCert.pubKey[3:35]),
// 		Y:     new(big.Int).SetBytes(pairStep1Resp.safecardCert.pubKey[35:67]),
// 	}

// 	pubKeyValid := validateECCPubKey(cardPubKey)
// 	log.Info("certificate public key valid: ", pubKeyValid)
// 	if !pubKeyValid {
// 		log.Error("card pubkey invalid")
// 		return err
// 	}

// 	secretHash, cryptogram, err := computeECDHSharedSecret(clientSalt, clientPrivKey, pairStep1Resp.safecardSalt, cardPubKey, pairStep1Resp.safecardSig)
// 	if err != nil {
// 		log.Error("could not compute shared secret. err: ", err)
// 		return err
// 	}
// 	//Pair Step 2
// 	pairStep2Resp, err := c.sendPairStep2(cryptogram)
// 	if err != nil {
// 		log.Error("error in pair step 2 command. err: ", err)
// 		return err
// 	}
// 	log.Infof("pairStep2Resp: % X", pairStep2Resp)

// 	//Derive Pairing Key
// 	pairingKey := sha256.Sum256(append(pairStep2Resp.salt, secretHash...))
// 	log.Infof("derived pairing key: % X", pairingKey)
// 	c.pairingKey = pairingKey[0:]
// 	return nil
// }

// func (c Safecard) Send(cmd *apdu.Command) ([]byte, error) {
// 	cmd.SetLe(0)

// 	rawCmd, err := cmd.Serialize()
// 	if err != nil {
// 		log.Error("could not serialize pair step 1 apdu. err: ", err)
// 		return nil, err
// 	}
// 	resp, err := c.card.Transmit(rawCmd)
// 	if err != nil {
// 		log.Error("could not send pair step 1 command. err: ", err)
// 		return nil, err
// 	}
// 	if len(resp) < 2 {
// 		return nil, errors.New("response was invalid length, too short")
// 	}
// 	// responseSuccessCode := []byte{0x90, 0x00}
// 	// if resp[len(resp)-2:] != responseSuccessCode {
// 	// 	return nil, errors.New("response success code was not received")
// 	// }
// 	return resp[:len(resp)-2], nil
// }

// func (c Safecard) OpenSecureChannel() error {
// 	privKey, err := secp256k1.GeneratePrivateKey()
// 	if err != nil {
// 		log.Error("could not generate secure channel private key. err: ", err)
// 	}
// 	apdu := NewAPDUOpenSecureChannel(privKey.PublicKey)
// 	resp, err := c.c.Send(apdu)
// 	if err != nil {
// 		log.Error("unable to send open secure channel command. err: ", err)
// 		return err
// 	}
// 	log.Infof("received secure channel response of length %v:\n% X", len(resp.Data), resp.Data)
// 	return nil

// }
