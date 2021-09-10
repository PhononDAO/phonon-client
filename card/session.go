package card

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/phonon-client/model"
)

/*The session struct handles a local connection with a card
Keeps a client side cache of the card state to make interaction
with the card through this API more convenient*/
type Session struct {
	cs          PhononCard
	active      bool
	initialized bool
	cardPaired  bool
}

var ErrAlreadyInitialized = errors.New("card is already initialized with a pin")
var ErrInitFailed = errors.New("card failed initialized check after init command accepted")
var ErrCardNotPairedToCard = errors.New("card not paired with any other card")

func NewSession(storage PhononCard, initialized bool) *Session {
	return &Session{
		cs:          storage,
		active:      true,
		initialized: initialized,
		cardPaired:  false,
	}
}

func NewSessionWithReaderIndex(index int) (*Session, error) {
	cs, initialized, err := OpenBestConnectionWithReaderIndex(index)
	if err != nil {
		return nil, err
	}
	return &Session{
		cs:          cs,
		active:      true,
		initialized: initialized,
	}, nil
}

func (s *Session) Init(pin string) error {
	if s.initialized {
		return ErrAlreadyInitialized
	}
	err := s.cs.Init(pin)
	if err != nil {
		return err
	}
	//Open new secure connection now that card is initialized

	err = s.cs.Pair()
	if err != nil {
		return err
	}
	err = s.cs.OpenSecureChannel()
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) VerifyPIN(pin string) error {
	err := s.cs.VerifyPIN(pin)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) CreatePhonon() (keyIndex uint16, pubkey *ecdsa.PublicKey, err error) {
	return s.cs.CreatePhonon()
}

func (s *Session) SetDescriptor(keyIndex uint16, currencyType model.CurrencyType, value float32) error {
	return s.cs.SetDescriptor(keyIndex, currencyType, value)
}

func (s *Session) ListPhonons(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) ([]model.Phonon, error) {
	return s.cs.ListPhonons(currencyType, lessThanValue, greaterThanValue)
}

func (s *Session) InitCardPairing() ([]byte, error) {
	return s.cs.InitCardPairing()
}

func (s *Session) CardPair(initPairingData []byte) ([]byte, error) {
	return s.cs.CardPair(initPairingData)
}

func (s *Session) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	cardPair2Data, err = s.cs.CardPair2(cardPairData)
	if err != nil {
		return nil, err
	}
	s.cardPaired = true
	return cardPair2Data, nil
}

func (s *Session) FinalizeCardPair(cardPair2Data []byte) error {
	err := s.cs.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	s.cardPaired = true
	return nil
}

func (s *Session) SendPhonons(keyIndices []uint16) ([]byte, error) {
	if !s.cardPaired {
		return nil, ErrCardNotPairedToCard
	}
	phononTransferPacket, err := s.cs.SendPhonons(keyIndices, false)
	if err != nil {
		return nil, err
	}

	return phononTransferPacket, nil
}

func (s *Session) ReceivePhonons(phononTransferPacket []byte) error {
	if !s.cardPaired {
		return ErrCardNotPairedToCard
	}
	err := s.cs.ReceivePhonons(phononTransferPacket)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) ReceiveInvoice(invoiceData []byte) error {
	if !s.cardPaired {
		return ErrCardNotPairedToCard
	}
	err := s.cs.ReceiveInvoice(invoiceData)
	if err != nil {
		return err
	}
	return nil
}
