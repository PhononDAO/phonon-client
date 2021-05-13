package card

import (
	"github.com/GridPlus/phonon-client/model"
)

type Session struct {
	cs     *PhononCommandSet
	active bool
}

func NewSession() (*Session, error) {
	cs, err := OpenBestConnection()
	if err != nil {
		return nil, err
	}
	return &Session{
		cs:     cs,
		active: true,
	}, nil
}

//TODO: fix this paradigm
// func (s *Session) checkActive() (*Session, error) {
// 	if !s.active {
// 		var err error
// 		if s, err = NewSession(); err != nil {
// 			return nil, err
// 		}
// 		return s, nil
// 	}
// 	return s, nil
// }

func (s *Session) ListPhonons(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) ([]model.Phonon, error) {
	if !s.active {
		var err error
		if s, err = NewSession(); err != nil {
			return nil, err
		}
	}
	phonons, err := s.cs.ListPhonons(currencyType, lessThanValue, greaterThanValue)
	if err != nil {
		return nil, err
	}
	//TODO: additional filtering options if necessary
	for _, phonon := range phonons {
		phonon.PubKey, err = s.cs.GetPhononPubKey(phonon.KeyIndex)
		if err != nil {
			return nil, err
		}
	}
	return phonons, nil
}

func SendPhonons() {
	//TODO implement
	return
}

func ReceivePhonons() {
	//TODO implement
	return
}
