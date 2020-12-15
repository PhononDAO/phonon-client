package card

import (
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
)

type MockCard struct {
	Phonons []model.Phonon
}

func (c MockCard) CreatePhonons(n int) (pubKeys [][]byte, err error) {
	phononPubKeys := make([][]byte, 0)
	for i := 0; i < n; i++ {
		phononPubKeys = append(phononPubKeys, util.RandomKey(32))
	}
	return phononPubKeys, nil
}

func (c MockCard) SetDescriptors(phonons []model.Phonon) error {
	c.Phonons = append(c.Phonons, phonons...)
	return nil
}

func (c MockCard) Select() error {
	//not implemented
	return nil
}
func (c MockCard) OpenChannel() (string, error) {
	//not implemented
	return "", nil
}

func (c MockCard) MutualAuthChannel() error {
	//not implemented
	return nil
}

func (c MockCard) VerifyPin() error {
	//not implemented
	return nil
}

func (c MockCard) ListPhonons(limit int, filterType string, filterValue []byte) (phonons []model.Phonon, err error) {
	numStoredPhonons := len(c.Phonons)
	if limit > numStoredPhonons {
		limit = numStoredPhonons
	}
	return phonons[0:limit], nil
}

func (c MockCard) SendPhonons(phononIDs []int) (transaction []byte, err error) {
	//not implemented
	return nil, nil
}

func (c MockCard) ReceivePhonons(transaction []byte) (err error) {
	//not implemented
	return nil
}
func (c MockCard) DestroyPhonon(phononID string) (err error) {
	//not implemented
	return nil
}
