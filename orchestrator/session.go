package orchestrator

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/chain"
	"github.com/GridPlus/phonon-client/model"
	remote "github.com/GridPlus/phonon-client/remote/v1/client"
	"github.com/GridPlus/phonon-client/util"

	log "github.com/sirupsen/logrus"
)

var ErrorRequestNotRecognized = errors.New("Unrecognized request sent to session")

/*The session struct handles a local connection with a card
Keeps a client side cache of the card state to make interaction
with the card through this API more convenient*/
type Session struct {
	cs                    model.PhononCard
	RemoteCard            model.CounterpartyPhononCard
	identityPubKey        *ecdsa.PublicKey
	remoteMessageChan     chan (model.SessionRequest)
	remoteMessageKillChan chan interface{}
	active                bool
	pinInitialized        bool
	terminalPaired        bool
	pinVerified           bool
	Cert                  *cert.CardCertificate
	ElementUsageMtex      sync.Mutex
}

var ErrAlreadyInitialized = errors.New("card is already initialized with a pin")
var ErrInitFailed = errors.New("card failed initialized check after init command accepted")
var ErrCardNotPairedToCard = errors.New("card not paired with any other card")

//Creates a new card session, automatically connecting if the card is already initialized with a PIN
//The next step is to run VerifyPIN to gain access to the secure commands on the card
func NewSession(storage model.PhononCard) (s *Session, err error) {
	s = &Session{
		cs:                    storage,
		active:                true,
		terminalPaired:        false,
		pinVerified:           false,
		ElementUsageMtex:      sync.Mutex{},
		remoteMessageChan:     make(chan model.SessionRequest),
		remoteMessageKillChan: make(chan interface{}),
	}
	s.ElementUsageMtex.Lock()
	_, _, s.pinInitialized, err = s.cs.Select()
	s.ElementUsageMtex.Unlock()
	if err != nil {
		log.Error("cannot select card for new session: ", err)
		return nil, err
	}
	if !s.pinInitialized {
		return s, nil
	}
	//If card is already initialized, go ahead and open terminal to card secure channel
	err = s.Connect()
	if err != nil {
		log.Error("could not run session connect: ", err)
		return nil, err
	}
	// launch session request handler
	go s.handleIncommingSessionRequests()
	return s, nil
}

// loop until killed
func (s *Session) handleIncommingSessionRequests() {
	for {
		select {
		case req := <-s.remoteMessageChan:
			s.handleRequest(req)
		case <-s.remoteMessageKillChan:
			return
		}
	}
}

func (s *Session) SetPaired(status bool) {
}

func (s *Session) GetName() string {
	if s.Cert == nil {
		return "unknown"
	}
	if s.Cert.PubKey != nil {
		return util.CardIDFromPubKey(s.identityPubKey)
	}
	return "unknown"
}
func (s *Session) GetCertificateSerialized() ([]byte, error) {
	cert, err := s.GetCertificate()
	if err != nil {
		return []byte{}, err
	} else {
		return cert.Serialize(), nil
	}
}
func (s *Session) GetCertificate() (*cert.CardCertificate, error) {
	if s.Cert != nil {
		log.Debugf("GetCertificate returning cert: % X", s.Cert)
		return s.Cert, nil
	}

	return &cert.CardCertificate{}, errors.New("certificate not cached by session yet")
}

func (s *Session) IsUnlocked() bool {

	return s.pinVerified
}

func (s *Session) IsInitialized() bool {
	return s.pinInitialized
}

func (s *Session) IsPairedToCard() bool {
	return s.RemoteCard != nil
}

//Connect opens a secure channel with a card.
func (s *Session) Connect() error {
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()
	cert, err := s.cs.Pair()
	if err != nil {
		return err
	}
	s.Cert = cert
	s.identityPubKey, _ = util.ParseECCPubKey(s.Cert.PubKey)
	err = s.cs.OpenSecureChannel()
	if err != nil {
		return err
	}
	s.terminalPaired = true
	return nil
}

//Initializes the card with a PIN
//Also creates a secure channel and verifies the PIN that was just set
func (s *Session) Init(pin string) error {
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

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
	s.pinVerified = true

	return nil
}

func (s *Session) VerifyPIN(pin string) error {
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	err := s.cs.VerifyPIN(pin)
	if err != nil {
		return err
	}
	s.pinVerified = true
	return nil
}

func (s *Session) ChangePIN(pin string) error {
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	if !s.pinVerified {
		return errors.New("card locked, cannot change pin")
	}
	return s.cs.ChangePIN(pin)
}

func (s *Session) verified() bool {
	if s.pinVerified && s.terminalPaired {
		return true
	}
	return false
}

func (s *Session) CreatePhonon() (keyIndex uint16, pubkey *ecdsa.PublicKey, err error) {
	if !s.verified() {
		return 0, nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.CreatePhonon(model.Secp256k1)
}

func (s *Session) SetDescriptor(p *model.Phonon) error {
	if !s.verified() {
		return card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()
	return s.cs.SetDescriptor(p)
}

func (s *Session) ListPhonons(currencyType model.CurrencyType, lessThanValue uint64, greaterThanValue uint64) ([]*model.Phonon, error) {
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.ListPhonons(currencyType, lessThanValue, greaterThanValue)
}

func (s *Session) GetPhononPubKey(keyIndex uint16) (pubkey *ecdsa.PublicKey, err error) {
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.GetPhononPubKey(keyIndex)
}

func (s *Session) DestroyPhonon(keyIndex uint16) (privKey *ecdsa.PrivateKey, err error) {
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.DestroyPhonon(keyIndex)
}

func (s *Session) IdentifyCard(nonce []byte) (cardPubKey *ecdsa.PublicKey, cardSig *util.ECDSASignature, err error) {
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.IdentifyCard(nonce)
}

func (s *Session) InitCardPairing(receiverCert cert.CardCertificate) ([]byte, error) {
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.InitCardPairing(receiverCert)
}

func (s *Session) CardPair(initPairingData []byte) ([]byte, error) {
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.CardPair(initPairingData)
}

func (s *Session) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	cardPair2Data, err = s.cs.CardPair2(cardPairData)
	if err != nil {
		return nil, err
	}
	log.Debug("set card session paired")
	return cardPair2Data, nil
}

func (s *Session) FinalizeCardPair(cardPair2Data []byte) error {
	if !s.verified() {
		return card.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	err := s.cs.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) SendPhonons(keyIndices []uint16) error {
	if !s.verified() && s.RemoteCard != nil {
		return ErrCardNotPairedToCard
	}
	err := s.RemoteCard.VerifyPaired()
	if err != nil {
		return err
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()
	phononTransferPacket, err := s.cs.SendPhonons(keyIndices, false)
	if err != nil {
		return err
	}
	err = s.RemoteCard.ReceivePhonons(phononTransferPacket)
	if err != nil {
		log.Debug("error receiving phonons on remote")
		return err
	}
	return nil
}

func (s *Session) ReceivePhonons(phononTransferPacket []byte) error {
	if !s.verified() && s.RemoteCard != nil {
		return ErrCardNotPairedToCard
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	err := s.cs.ReceivePhonons(phononTransferPacket)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) GenerateInvoice() ([]byte, error) {
	if !s.verified() && s.RemoteCard != nil {
		return nil, ErrCardNotPairedToCard
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.GenerateInvoice()
}

func (s *Session) ReceiveInvoice(invoiceData []byte) error {
	if !s.verified() && s.RemoteCard != nil {
		return ErrCardNotPairedToCard
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	err := s.cs.ReceiveInvoice(invoiceData)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) ConnectToRemoteProvider(RemoteURL string) error {
	u, err := url.Parse(RemoteURL)
	if err != nil {
		return fmt.Errorf("unable to parse url for card connection: %s", err.Error())
	}
	log.Info("connecting")
	//guard
	remConn, err := remote.Connect(s.remoteMessageChan, fmt.Sprintf("https://%s/phonon", u.Host), true)
	if err != nil {
		return fmt.Errorf("unable to connect to remote session: %s", err.Error())
	}
	s.RemoteCard = remConn
	return nil
}

func (s *Session) ConnectToLocalProvider() error {
	// uhh
	return nil
}

func (s *Session) ConnectToCounterparty(cardID string) error {
	err := s.RemoteCard.ConnectToCard(cardID)
	if err != nil {
		log.Info("returning error from ConnectRemoteSession")
		return err
	}
	localPubKey, err := util.ParseECCPubKey(s.Cert.PubKey)
	if err != nil {
		//we shouldn't get this far and still receive this error
		return err
	}
	if cardID < util.CardIDFromPubKey(localPubKey) {
		paired := make(chan bool, 1)
		go func() {
			for {
				if s.IsPairedToCard() {
					paired <- true
				}
				time.Sleep(10 * time.Millisecond)
			}
		}()
		select {
		case <-time.After(30 * time.Second):
			return errors.New("pairing timed out")
		case <-paired:
			return nil
		}
	}
	err = s.PairWithRemoteCard(s.RemoteCard)
	return err

}

func (s *Session) PairWithRemoteCard(remoteCard model.CounterpartyPhononCard) error {
	remoteCert, err := remoteCard.GetCertificate()
	if err != nil {
		return err
	}
	initPairingData, err := s.InitCardPairing(*remoteCert)
	if err != nil {
		return err
	}
	log.Debug("sending card pair request")
	cardPairData, err := remoteCard.CardPair(initPairingData)
	if err != nil {
		return err
	}
	cardPair2Data, err := s.CardPair2(cardPairData)
	if err != nil {
		log.Debug("PairWithRemoteCard failed at cardPair2. err: ", err)
		return err
	}
	err = remoteCard.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	s.RemoteCard = remoteCard
	return nil
}

/*InitDepositPhonons takes a currencyType and a map of denominations to quantity,
Creates the required phonons, deposits them using the configured service for the asset
and upon success sets their descriptors*/
func (s *Session) InitDepositPhonons(currencyType model.CurrencyType, denoms []model.Denomination) (phonons []*model.Phonon, err error) {
	log.Debugf("running InitDepositPhonons with data: %v, %v\n", currencyType, denoms)
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	for _, denom := range denoms {
		p := &model.Phonon{}
		p.KeyIndex, p.PubKey, err = s.CreatePhonon()
		log.Debug("ran CreatePhonons in InitDepositLoop")
		if err != nil {
			log.Error("failed to create phonon for deposit: ", err)
			return nil, err
		}
		p.Denomination = denom
		p.CurrencyType = currencyType
		p.Address, err = chain.DeriveAddress(p)
		if err != nil {
			log.Error("failed to derive address for phonon deposit: ", err)
			return nil, err
		}

		phonons = append(phonons, p)
	}
	return phonons, nil
}

type DepositConfirmation struct {
	Phonon           *model.Phonon
	ConfirmedOnChain bool
	ConfirmedOnCard  bool
}

func (s *Session) FinalizeDepositPhonons(confirmations []DepositConfirmation) ([]DepositConfirmation, error) {
	log.Debug("running finalizeDepositPhonon")
	if !s.verified() {
		return nil, card.ErrPINNotEntered
	}
	var lastErr error
	for _, v := range confirmations {
		err := s.FinalizeDepositPhonon(v)
		if err != nil {
			lastErr = err
			v.ConfirmedOnCard = false
		} else {
			v.ConfirmedOnCard = true
		}
	}
	return confirmations, lastErr
}

func (s *Session) FinalizeDepositPhonon(dc DepositConfirmation) error {
	if dc.ConfirmedOnChain {

		err := s.SetDescriptor(dc.Phonon)
		if err != nil {
			log.Error("unable to finalize deposit by setting descriptor for phonon: ", dc.Phonon)
			return err
		}
	} else {
		_, err := s.DestroyPhonon(dc.Phonon.KeyIndex)
		if err != nil {
			log.Error("unable to clean up deposit failure by destroying phonon: ", dc.Phonon)
		}
	}
	return nil
}

// the panics are paths that should NEVER be found in runtime as it's already been determined by the case statement.
func (s *Session) handleRequest(r model.SessionRequest) {
	switch r.GetName() {
	case "RequestCertificate":
		req, ok := r.(*model.RequestCertificate)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseCertificate
		resp.Payload, resp.Err = s.GetCertificate()
		req.Ret <- resp
	case "RequestIdentifyCard":
		req, ok := r.(*model.RequestIdentifyCard)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseIdentifyCard
		resp.PubKey, resp.Sig, resp.Err = s.IdentifyCard(req.Nonce)
		req.Ret <- resp
	case "RequestCardPair1":
		req, ok := r.(*model.RequestCardPair1)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseCardPair1
		resp.Payload, resp.Err = s.CardPair(req.Payload)
		req.Ret <- resp
	case "RequestFinalizeCardPair":
		req, ok := r.(*model.RequestFinalizeCardPair)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseFinalizeCardPair
		resp.Err = s.FinalizeCardPair(req.Payload)
		req.Ret <- resp

	case "RequestSetRemote":
		req, ok := r.(*model.RequestSetRemote)
		if !ok {
			panic("this shouldn't happen.")
		}
		s.RemoteCard = req.Card
		var resp model.ResponseSetRemote
		resp.Err = nil
		req.Ret <- resp
	case "RequestReceivePhonons":
		req, ok := r.(*model.RequestReceivePhonons)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseReceivePhonons
		resp.Err = s.ReceivePhonons(req.Payload)
		req.Ret <- resp
	case "RequestGetName":
		req, ok := r.(*model.RequestGetName)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseGetName
		resp.Name = s.GetName()
		resp.Err = nil
		req.Ret <- resp
	case "RequestPairWithRemote":
		req, ok := r.(*model.RequestPairWithRemote)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponsePairWithRemote
		resp.Err = s.PairWithRemoteCard(req.Card)
		req.Ret <- resp
	case "RequestSetPaired":
		req, ok := r.(*model.RequestSetPaired)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponseSetPaired
		s.SetPaired(req.Status)
		resp.Err = nil
		req.Ret <- resp
	}
}
