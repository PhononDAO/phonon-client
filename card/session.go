package card

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	log "github.com/sirupsen/logrus"
)

/*The session struct handles a local connection with a card
Keeps a client side cache of the card state to make interaction
with the card through this API more convenient*/
type Session struct {
	cs             PhononCard
	identityPubKey *ecdsa.PublicKey
	active         bool
	pinInitialized bool
	terminalPaired bool
	pinVerified    bool
	cardPaired     bool
	cert           cert.CardCertificate
	name           string
}

var ErrAlreadyInitialized = errors.New("card is already initialized with a pin")
var ErrInitFailed = errors.New("card failed initialized check after init command accepted")
var ErrCardNotPairedToCard = errors.New("card not paired with any other card")

//Creates a new card session, automatically connecting if the card is already initialized with a PIN
//The next step is to run VerifyPIN to gain access to the secure commands on the card
func NewSession(storage PhononCard) (s *Session, err error) {
	s = &Session{
		cs:             storage,
		active:         true,
		terminalPaired: false,
		cardPaired:     false,
		pinVerified:    false,
	}
	_, _, s.pinInitialized, err = s.cs.Select()
	if err != nil {
		return nil, err
	}
	if !s.pinInitialized {
		return s, nil
	}
	//If card is already initialized, go ahead and open terminal to card secure channel
	err = s.Connect()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Session) GetName() string {
	//TODO: Use future card GET_NAME
	if s.cert.PubKey != nil {
		return util.ECDSAPubKeyToHexString(s.identityPubKey)
	}
	return "unknown"
}

func (s *Session) GetCertificate() (cert.CardCertificate, error) {
	//If s.Cert is already populated, return it
	if s.cert.PubKey != nil {
		log.Debugf("GetCertificate returning cert: % X", s.cert)
		return s.cert, nil
	}

	//TODO, fetch this if it's not there yet
	return cert.CardCertificate{}, errors.New("certificate not cached by session yet")
}

//Connect opens a secure channel with a card.
func (s *Session) Connect() error {
	cert, err := s.cs.Pair()
	if err != nil {
		return err
	}
	s.cert = cert
	s.identityPubKey, _ = util.ParseECDSAPubKey(s.cert.PubKey)
	err = s.cs.OpenSecureChannel()
	if err != nil {
		return err
	}
	s.terminalPaired = true
	return nil
}

//Initializes the card with a PIN
//Also creates a secure channel and verifies the PIN that was just set
//TODO: Fix MUTUAL_AUTH Error returned when called this way
func (s *Session) Init(pin string) error {
	if s.pinInitialized {
		return ErrAlreadyInitialized
	}
	err := s.cs.Init(pin)
	if err != nil {
		return err
	}
	s.pinInitialized = true
	//Open new secure connection now that card is initialized
	err = s.Connect()
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
	s.pinVerified = true
	return nil
}

func (s *Session) verified() bool {
	if s.pinVerified && s.terminalPaired {
		return true
	}
	return false
}

func (s *Session) CreatePhonon() (keyIndex uint16, pubkey *ecdsa.PublicKey, err error) {
	if !s.verified() {
		return 0, nil, ErrPINNotEntered
	}
	return s.cs.CreatePhonon()
}

func (s *Session) SetDescriptor(keyIndex uint16, currencyType model.CurrencyType, value float32) error {
	if !s.verified() {
		return ErrPINNotEntered
	}
	return s.cs.SetDescriptor(keyIndex, currencyType, value)
}

func (s *Session) ListPhonons(currencyType model.CurrencyType, lessThanValue float32, greaterThanValue float32) ([]model.Phonon, error) {
	if !s.verified() {
		return nil, ErrPINNotEntered
	}
	return s.cs.ListPhonons(currencyType, lessThanValue, greaterThanValue)
}

func (s *Session) InitCardPairing(receiverCert cert.CardCertificate) ([]byte, error) {
	if !s.verified() {
		return nil, ErrPINNotEntered
	}
	return s.cs.InitCardPairing(receiverCert)
}

func (s *Session) CardPair(initPairingData []byte) ([]byte, error) {
	if !s.verified() {
		return nil, ErrPINNotEntered
	}
	return s.cs.CardPair(initPairingData)
}

func (s *Session) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	if !s.verified() {
		return nil, ErrPINNotEntered
	}
	cardPair2Data, err = s.cs.CardPair2(cardPairData)
	if err != nil {
		return nil, err
	}
	s.cardPaired = true
	log.Debug("set card session paired")
	return cardPair2Data, nil
}

func (s *Session) FinalizeCardPair(cardPair2Data []byte) error {
	if !s.verified() {
		return ErrPINNotEntered
	}
	err := s.cs.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	s.cardPaired = true
	log.Debug("set card session paired")
	return nil
}

func (s *Session) SendPhonons(keyIndices []uint16) ([]byte, error) {
	if !s.verified() && !s.cardPaired {
		return nil, ErrCardNotPairedToCard
	}
	phononTransferPacket, err := s.cs.SendPhonons(keyIndices, false)
	if err != nil {
		return nil, err
	}

	return phononTransferPacket, nil
}

func (s *Session) ReceivePhonons(phononTransferPacket []byte) error {
	if !s.verified() && !s.cardPaired {
		return ErrCardNotPairedToCard
	}
	err := s.cs.ReceivePhonons(phononTransferPacket)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) GenerateInvoice() ([]byte, error) {
	if !s.verified() && !s.cardPaired {
		return nil, ErrCardNotPairedToCard
	}
	return s.cs.GenerateInvoice()
}

func (s *Session) ReceiveInvoice(invoiceData []byte) error {
	if !s.verified() && !s.cardPaired {
		return ErrCardNotPairedToCard
	}
	err := s.cs.ReceiveInvoice(invoiceData)
	if err != nil {
		return err
	}
	return nil
}
