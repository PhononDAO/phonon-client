package card

import (
	"errors"
	"github.com/GridPlus/phonon-client/model"
)

type Session struct {
	cs          *PhononCommandSet
	active      bool
	initialized bool
}

var ErrAlreadyInitialized = errors.New("card is already initialized with a pin")
var ErrInitFailed = errors.New("card failed initialized check after init command accepted")

func NewSession() (*Session, error) {
	cs, initialized, err := OpenBestConnection()
	if err != nil {
		return nil, err
	}
	return &Session{
		cs:          cs,
		active:      true,
		initialized: initialized,
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

func (s *Session) Init(pin string) error {
	if s.initialized {
		return ErrAlreadyInitialized
	}
	err := s.cs.Init(pin)
	if err != nil {
		return err
	}
	//Open new secure connection now that card is initialized
	var initialized bool
	s.cs, initialized, err = OpenBestConnection()
	if err != nil {
		return err
	}
	if !initialized {
		return ErrInitFailed
	}
	return nil
}

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

//TODO: Genericize for generic KV Pairs
// func (s *Session) DepositPhonon(currencyType model.CurrencyType, value float32) (phonon model.Phonon, err error) {
// 	phonon.KeyIndex, phonon.PubKey, err = s.cs.CreatePhonon()
// 	if err != nil {
// 		return
// 	}
// }

func (s *Session) PairWithRemoteCard(remoteCard model.RemotePhononCard) error {
	initPairingData, err := s.cs.InitCardPairing()
	if err != nil {
		return err
	}
	cardPairData, err := remoteCard.CardPair(initPairingData)
	if err != nil {
		return err
	}
	cardPair2Data, err := s.cs.CardPair2(cardPairData)
	if err != nil {
		return err
	}
	err = remoteCard.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) SendPhonons(remoteCard model.RemotePhononCard) {

}

func ReceivePhonons() {
	//TODO implement
	return
}
