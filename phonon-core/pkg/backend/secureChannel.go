package backend

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"strings"

	"github.com/GridPlus/keycard-go/crypto"
	"github.com/GridPlus/keycard-go/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var ErrInvalidResponseMAC = errors.New("invalid response MAC")

type SecureChannel struct {
	c          types.Channel
	open       bool
	secret     []byte
	publicKey  *ecdsa.PublicKey
	encKey     []byte
	macKey     []byte
	iv         []byte
	ApduLogger logrus.Logger
}

func NewSecureChannel(c types.Channel, logger logrus.Logger) *SecureChannel {
	return &SecureChannel{
		c:          c,
		ApduLogger: logger,
	}
}

// the following are temoporary
func (sc *SecureChannel) Open() bool {
	return sc.open
}

func (sc *SecureChannel) Iv() []byte {
	return sc.iv
}

func (sc *SecureChannel) EncKey() []byte {
	return sc.encKey
}

func (sc *SecureChannel) C() types.Channel {
	return sc.c
}

// end temporary functions

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
func (sc *SecureChannel) UpdateIV(meta, data []byte) error {
	mac, err := crypto.CalculateMac(meta, data, sc.macKey)
	if err != nil {
		return err
	}
	sc.iv = mac
	return nil
}

// Encrypt data and return ciphertext directly
func (sc *SecureChannel) Encrypt(data []byte) (ciphertext []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("recovered from panic: ", r)
			err = errors.New("recovered from panic in secure channel send")
		}
	}()
	encData, err := crypto.EncryptData(data, sc.encKey, sc.iv)
	if err != nil {
		return nil, err
	}
	//Not sure if the format of meta matters, seems to me it is just arbitrary data of length 16
	//Not even sure what the point of meta is, seems like it could be removed with no consequences, but I might
	//be missing something
	// meta := []byte{0, 0, 0, 0, byte(len(encData) + 16), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	// meta := util.RandomKey(16)
	meta := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if err = sc.UpdateIV(meta, encData); err != nil {
		return nil, err
	}
	ciphertext = append(sc.iv, encData...)

	log.Debugf("encrypting this card to card: % X\n", ciphertext)
	return ciphertext, nil
}

// DecryptDirect decrypts a message but does not track iv updates or authenticate the decryption with the MAC
// Useful for decrypting a message that was just encrypted by the same channel, rather than by a counterparty channel
// which is keeping the IV in sync. Could also be used to decrypt a message which provides the IV
func (sc *SecureChannel) DecryptDirect(ciphertext []byte, iv []byte) (data []byte, err error) {
	log.Debugf("ciphertext: % X", ciphertext)
	// meta := []byte{byte(len(ciphertext)), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	// mac := ciphertext[:len(sc.iv)]
	// iv := ciphertext[:len(sc.iv)]
	encData := ciphertext[len(sc.iv):]
	//Use sc.iv if a counteparty has encrypted a response to the last encrypted message
	data, err = crypto.DecryptData(encData, sc.encKey, iv)
	if err != nil {
		return nil, err
	}
	// if err = sc.updateIV(meta, encData); err != nil {
	// 	return nil, err
	// }
	// if !bytes.Equal(sc.iv, mac) {
	// 	return nil, ErrInvalidResponseMAC
	// }
	return data, nil
}

// Decrypts the response to the last message Encrypted in this channel
// The init vector is automatically updated to match the iv the response should use after
// a Decrypt -> Encrypt cycle
func (sc *SecureChannel) Decrypt(ciphertext []byte) (data []byte, err error) {
	log.Debug("decrypting ciphertext of length: ", len(ciphertext))
	log.Debugf("% X", ciphertext)
	//MAC is prepended, should be equal to last seen init vector
	mac := ciphertext[:len(sc.iv)]
	//encrypted data follows MAC
	encData := ciphertext[len(sc.iv):]
	data, err = crypto.DecryptData(encData, sc.encKey, sc.iv)
	if err != nil {
		return nil, err
	}
	//Meta is predetermined 16 byte value to provide initial cipher block for mac calculation
	meta := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	log.Debugf("meta: % X", meta)
	log.Debugf("data: % X", data)
	if err = sc.UpdateIV(meta, encData); err != nil {
		return nil, err
	}
	//Matching MAC confirms that the ciphertext was decrypted correctly
	//and that the channel iv's are in sync
	if !bytes.Equal(sc.iv, mac) {
		return nil, ErrInvalidResponseMAC
	}
	return data, nil
}
