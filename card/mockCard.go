package card

import (
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type MockCard struct {
	Phonons []model.Phonon
	pin     string
	sc      SecureChannel
}

//TODO
func (c MockCard) Select() (instanceUID []byte, cardPubKey []byte, cardInitialized bool, err error) {
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

//TODO
func (c MockCard) CreatePhonons(n int) (pubKeys [][]byte, err error) {
	phononPubKeys := make([][]byte, 0)
	for i := 0; i < n; i++ {
		//65 bytes ECC key
		phononPubKeys = append(phononPubKeys, util.RandomKey(65))
	}
	return phononPubKeys, nil
}

//TODO
func (c MockCard) SetDescriptors(phonons []model.Phonon) error {
	c.Phonons = append(c.Phonons, phonons...)
	return nil
}

//TODO
func (c MockCard) OpenChannel() (string, error) {
	//not implemented
	return "", nil
}

//TODO
func (c MockCard) MutualAuthChannel() error {
	//not implemented
	return nil
}

//TODO
func (c MockCard) VerifyPin() error {
	//not implemented
	return nil
}

//TODO
func (c MockCard) ListPhonons(limit int, filterType string, filterValue []byte) (phonons []model.Phonon, err error) {
	numStoredPhonons := len(c.Phonons)
	if limit > numStoredPhonons {
		limit = numStoredPhonons
	}
	return phonons[0:limit], nil
}

//TODO
func (c MockCard) SendPhonons(phononIDs []int) (transaction []byte, err error) {
	//not implemented
	return nil, nil
}

//TODO
func (c MockCard) ReceivePhonons(transaction []byte) (err error) {
	//not implemented
	return nil
}

//TODO
func (c MockCard) DestroyPhonon(phononID string) (err error) {
	//not implemented
	return nil
}
