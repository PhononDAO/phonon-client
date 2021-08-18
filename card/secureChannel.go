package card

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"strings"

	"github.com/GridPlus/keycard-go/apdu"
	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/globalplatform"
	"github.com/GridPlus/keycard-go/hexutils"
	"github.com/GridPlus/keycard-go/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

var ErrInvalidResponseMAC = errors.New("invalid response MAC")

type SecureChannel struct {
	c         types.Channel
	open      bool
	secret    []byte
	publicKey *ecdsa.PublicKey
	encKey    []byte
	macKey    []byte
	iv        []byte
}

func NewSecureChannel(c types.Channel) *SecureChannel {
	return &SecureChannel{
		c: c,
	}
}

func (sc *SecureChannel) GenerateSecret(cardPubKeyData []byte) error {
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		return err
	}

	cardPubKey, err := ethcrypto.UnmarshalPubkey(cardPubKeyData)
	if err != nil {
		return err
	}

	sc.publicKey = &key.PublicKey
	sc.secret = crypto.GenerateECDHSharedSecret(key, cardPubKey)

	return nil
}

func (sc *SecureChannel) GenerateStaticSecret(cardPubKeyData []byte) error {
	//Generate a static 40 byte value suitable for generating a predictable key
	var seed string
	for i := 0; i < 40; i++ {
		seed += "A"
	}
	staticSeed := strings.NewReader(seed)
	key, err := ecdsa.GenerateKey(ethcrypto.S256(), staticSeed)
	if err != nil {
		return err
	}
	cardPubKey, err := ethcrypto.UnmarshalPubkey(cardPubKeyData)
	if err != nil {
		return err
	}

	sc.publicKey = &key.PublicKey
	sc.secret = crypto.GenerateECDHSharedSecret(key, cardPubKey)

	return nil
}

func (sc *SecureChannel) Reset() {
	sc.open = false
}

func (sc *SecureChannel) Init(iv, encKey, macKey []byte) {
	sc.iv = iv
	sc.encKey = encKey
	sc.macKey = macKey
	sc.open = true
}

func (sc *SecureChannel) Secret() []byte {
	return sc.secret
}

func (sc *SecureChannel) PublicKey() *ecdsa.PublicKey {
	return sc.publicKey
}

func (sc *SecureChannel) RawPublicKey() []byte {
	return ethcrypto.FromECDSAPub(sc.publicKey)
}

//AES-CBC-256 Symmetric encryption
func (sc *SecureChannel) Send(cmd *Command) (resp *apdu.Response, err error) {
	log.Debugf("raw command before encryption: CLA: % X Ins: % X P1: % X P2: % X Data: % X", cmd.ApduCmd.Cla, cmd.ApduCmd.Ins, cmd.ApduCmd.P1, cmd.ApduCmd.P2, cmd.ApduCmd.Data)
	defer func() {
		if r := recover(); r != nil {
			log.Error("recovered from panic: ", r)
			err = errors.New("recovered from panic in secure channel send")
		}
	}()
	if sc.open {
		encData, err := crypto.EncryptData(cmd.ApduCmd.Data, sc.encKey, sc.iv)
		if err != nil {
			return nil, err
		}

		meta := []byte{cmd.ApduCmd.Cla, cmd.ApduCmd.Ins, cmd.ApduCmd.P1, cmd.ApduCmd.P2, byte(len(encData) + 16), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		if err = sc.updateIV(meta, encData); err != nil {
			return nil, err
		}

		newData := append(sc.iv, encData...)
		cmd.ApduCmd.Data = newData
	}

	resp, err = sc.c.Send(cmd.ApduCmd)
	if err != nil {
		return nil, err
	}

	if resp.Sw != globalplatform.SwOK {
		return nil, apdu.NewErrBadResponse(resp.Sw, "unexpected sw in secure channel")
	}

	rmeta := []byte{byte(len(resp.Data)), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rmac := resp.Data[:len(sc.iv)]
	rdata := resp.Data[len(sc.iv):]
	plainData, err := crypto.DecryptData(rdata, sc.encKey, sc.iv)
	if err != nil {
		return nil, err
	}
	if err = sc.updateIV(rmeta, rdata); err != nil {
		return nil, err
	}

	if !bytes.Equal(sc.iv, rmac) {
		return nil, ErrInvalidResponseMAC
	}

	log.Debug("apdu response decrypted hex: ", hexutils.BytesToHexWithSpaces(plainData))

	return ParseResponseWithErrCheck(cmd, plainData)
}

func ParseResponseWithErrCheck(cmd *Command, plainData []byte) (*apdu.Response, error) {
	res, err := apdu.ParseResponse(plainData)
	if err != nil {
		return res, err
	}
	err = cmd.HumanReadableErr(res)
	return res, err
}

func (sc *SecureChannel) updateIV(meta, data []byte) error {
	mac, err := crypto.CalculateMac(meta, data, sc.macKey)
	if err != nil {
		return err
	}
	sc.iv = mac
	return nil
}

//TODO: Make sure I can delete this
// func (sc *SecureChannel) OneShotEncrypt(secrets *Secrets) ([]byte, error) {
// 	pubKeyData := ethcrypto.FromECDSAPub(sc.publicKey)
// 	data := append([]byte(secrets.Pin()), []byte(secrets.Puk())...)
// 	data = append(data, secrets.PairingToken()...)

// 	return crypto.OneShotEncrypt(pubKeyData, sc.secret, data)
// }
