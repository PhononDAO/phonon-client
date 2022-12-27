package orchestrator

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/GridPlus/phonon-core/internal/util"
	"github.com/GridPlus/phonon-core/pkg/backend"
	card "github.com/GridPlus/phonon-core/pkg/backend"
	"github.com/GridPlus/phonon-core/pkg/cert"
	"github.com/GridPlus/phonon-core/pkg/chain"
	"github.com/GridPlus/phonon-core/pkg/model"
	remote "github.com/GridPlus/phonon-core/pkg/remote/v1/client"

	log "github.com/sirupsen/logrus"
)

/*
The session struct handles a local connection with a card
Keeps a client side cache of the card state to make interaction
with the card through this API more convenient
*/
type Session struct {
	cs                    model.PhononCard
	RemoteCard            model.CounterpartyPhononCard
	identityPubKey        *ecdsa.PublicKey
	friendlyName          string
	remoteMessageChan     chan (model.SessionRequest)
	remoteMessageKillChan chan interface{}
	active                bool
	pinInitialized        bool
	terminalPaired        bool
	pinVerified           bool
	Cert                  *cert.CardCertificate
	ElementUsageMtex      sync.Mutex
	logger                *log.Entry
	chainSrv              chain.ChainService
	cache                 map[model.PhononKeyIndex]cachedPhonon
	cancelMiningChan      chan struct{}
	isMiningActive        bool
	mutexedMiningReport   mutexedMiningReport
	// cachePopulated indicates if all of the phonons present on the card have been cached. This is currently only set when listphonons is called with the values to list all phonons on the card.
	cachePopulated bool
}

const (
	StatusMiningSuccess   = "success"
	StatusMiningActive    = "active"
	StatusMiningError     = "error"
	StatusMiningCancelled = "cancelled"
)

type mutexedMiningReport struct {
	m    map[string]miningStatusReport
	mtex *sync.Mutex
}

type miningStatusReport struct {
	Attempts    int
	Status      string
	TimeElapsed int64
	StartTime   time.Time
	StopTime    time.Time `json:",omitempty"`
	AverageTime int64     `json:",omitempty"`
	KeyIndex    int       `json:",omitempty"`
	Hash        string    `json:",omitempty"`
}

type cachedPhonon struct {
	pubkeyCached bool
	infoCached   bool
	p            *model.Phonon
}

var ErrAlreadyInitialized = errors.New("card is already initialized with a pin")
var ErrInitFailed = errors.New("card failed initialized check after init command accepted")
var ErrCardNotPairedToCard = errors.New("card not paired with any other card")
var ErrNameCannotBeEmpty = errors.New("requested name cannot be empty")
var ErrMiningNotActive = errors.New("no active mining operation")
var ErrMiningReportNotAvailable = errors.New("could not find mining status report")

// Creates a new card session, automatically connecting if the card is already initialized with a PIN
// The next step is to run VerifyPIN to gain access to the secure commands on the card
func NewSession(storage model.PhononCard) (s *Session, err error) {
	chainSrv, err := chain.NewMultiChainRouter()
	if err != nil {
		return nil, err
	}
	s = &Session{
		cs:                    storage,
		RemoteCard:            nil,
		identityPubKey:        nil,
		friendlyName:          "",
		remoteMessageChan:     make(chan model.SessionRequest),
		remoteMessageKillChan: make(chan interface{}),
		active:                true,
		pinInitialized:        false,
		terminalPaired:        false,
		pinVerified:           false,
		Cert:                  nil,
		ElementUsageMtex:      sync.Mutex{},
		logger:                log.WithField("CardID", "unknown"),
		chainSrv:              chainSrv,
		cancelMiningChan:      make(chan struct{}),
		isMiningActive:        false,
		mutexedMiningReport:   mutexedMiningReport{m: make(map[string]miningStatusReport), mtex: &sync.Mutex{}},
		cache:                 make(map[model.PhononKeyIndex]cachedPhonon),
	}
	s.logger = log.WithField("cardID", s.GetCardId())

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
	go s.handleIncomingSessionRequests()
	log.Debug("initialized new applet session")
	return s, nil
}

func (m mutexedMiningReport) setMiningStatus(attemptId string, report miningStatusReport) {
	m.mtex.Lock()
	m.m[attemptId] = report
	m.mtex.Unlock()
}

func (m mutexedMiningReport) getMiningStatus(attemptId string) (miningStatusReport, error) {
	m.mtex.Lock()
	defer m.mtex.Unlock()
	if _, ok := m.m[attemptId]; ok {
		return m.m[attemptId], nil
	}
	return miningStatusReport{}, ErrMiningReportNotAvailable
}

// loop until killed
func (s *Session) handleIncomingSessionRequests() {
	for {
		select {
		case req := <-s.remoteMessageChan:
			s.handleRequest(req)
		case <-s.remoteMessageKillChan:
			return
		}
	}
}

func (s *Session) SetPaired(status bool) {}

func generateId() (string, error) {
	buffer := make([]byte, 16)
	_, err := rand.Read(buffer)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buffer), nil
}

func (s *Session) GetMiningReport(attemptId string) (miningStatusReport, error) {
	if !s.verified() {
		return miningStatusReport{}, backend.ErrPINNotEntered
	}

	return s.mutexedMiningReport.getMiningStatus(attemptId)
}

func (s *Session) ListMiningReports() (map[string]miningStatusReport, error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}

	if len(s.mutexedMiningReport.m) > 0 {
		return s.mutexedMiningReport.m, nil
	}

	return nil, ErrMiningReportNotAvailable
}

func (s *Session) CancelMiningRequest() error {
	if !s.verified() {
		return backend.ErrPINNotEntered
	}

	if s.isMiningActive {
		s.cancelMiningChan <- struct{}{}
		return nil
	}

	return ErrMiningNotActive
}

func (s *Session) initMiningAttempt(id string, difficulty uint8) {
	cancel := make(chan struct{})
	s.cancelMiningChan = cancel

	i := 0
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	report := miningStatusReport{}

	start := time.Now()
	for {
		select {
		case <-cancel:
			s.isMiningActive = false
			elapsed := time.Since(start)
			log.Debug("mining successfully cancelled")
			report.Attempts = i
			report.Status = StatusMiningCancelled
			report.TimeElapsed = elapsed.Milliseconds()
			report.StartTime = start
			report.StopTime = time.Now()
			s.mutexedMiningReport.setMiningStatus(id, report)
			return
		default:
			s.isMiningActive = true
			elapsed := time.Since(start)
			averageTime := time.Duration(float64(elapsed.Nanoseconds()) / float64(i))
			keyIndex, hash, err := s.cs.MineNativePhonon(difficulty)
			if err == card.ErrMiningFailed {
				log.Debug("mining failed to find a phonon, retrying...")
				report.Attempts = i
				report.Status = StatusMiningActive
				report.TimeElapsed = elapsed.Milliseconds()
				report.StartTime = start
				s.mutexedMiningReport.setMiningStatus(id, report)
			} else if err != nil {
				log.Debug("mining failed due to an unknown error: ", err)
				report.Attempts = i
				report.Status = StatusMiningError
				report.TimeElapsed = elapsed.Milliseconds()
				report.StartTime = start
				report.StopTime = time.Now()
				s.mutexedMiningReport.setMiningStatus(id, report)
				return
			} else {
				log.Debugf("mining succeeded with difficulty %d", difficulty)
				report.Attempts = i
				report.Status = StatusMiningSuccess
				report.TimeElapsed = elapsed.Milliseconds()
				report.StartTime = start
				report.StopTime = time.Now()
				report.AverageTime = averageTime.Milliseconds()
				report.KeyIndex = int(keyIndex)
				report.Hash = hex.EncodeToString(hash)
				s.mutexedMiningReport.setMiningStatus(id, report)
				phonons, err := s.cs.ListPhonons(0, 0, 0, false)
				if err != nil {
					log.Error("error listing phonons: ", err)
					return
				}
				for _, p := range phonons {
					if p.KeyIndex == keyIndex {
						pubkey, err := model.NewPhononPubKey(hash, model.NativeCurve)
						if err != nil {
							fmt.Println("error getting public key: ", err)
							return
						}

						p.PubKey = pubkey

						s.cache[keyIndex] = cachedPhonon{
							pubkeyCached: true,
							infoCached:   true,
							p:            p,
						}
					}
				}
				return
			}
			i += 1
		}
	}
}

func (s *Session) MineNativePhonon(difficulty uint8) (string, error) {
	if !s.verified() {
		return "", backend.ErrPINNotEntered
	}

	id, err := generateId()
	if err != nil {
		return "", err
	}

	go s.initMiningAttempt(id, difficulty)

	return id, nil
}

func (s *Session) GetCardId() string {
	//If identity public key has already been cached by pairing, return it
	if s.identityPubKey != nil {
		return util.CardIDFromPubKey(s.identityPubKey)
	} else {
		//else fetch identity public key directly through identify card
		pubKey, _, err := s.IdentifyCard(util.RandomKey(32))
		if err != nil {
			log.Error("error identifying card via GetName(). err: ", err)
			return "unknown"
		}
		s.identityPubKey = pubKey
		return util.CardIDFromPubKey(s.identityPubKey)
	}
}

func (s *Session) GetName() (string, error) {
	if s.friendlyName != "" {
		return s.friendlyName, nil
	} else {
		s.ElementUsageMtex.Lock()
		defer s.ElementUsageMtex.Unlock()
		var err error
		s.friendlyName, err = s.cs.GetFriendlyName()
		if err != nil {
			return "", err
		}
	}
	return s.friendlyName, nil
}

func (s *Session) SetName(name string) error {
	if !s.verified() {
		return backend.ErrPINNotEntered
	}
	if name == "" {
		return ErrNameCannotBeEmpty
	}
	err := s.cs.SetFriendlyName(name)
	if err == nil {
		s.friendlyName = name
	}
	return err
}

func (s *Session) GetCertificate() (*cert.CardCertificate, error) {
	if s.Cert != nil {
		log.Debugf("GetCertificate returning cert: %v", s.Cert)
		return s.Cert, nil
	}

	return &cert.CardCertificate{}, errors.New("certificate not cached by session yet")
}

func (s *Session) IsUnlocked() bool {
	return s.pinVerified
}

func (s *Session) IsPairedToTerminal() bool {
	return s.terminalPaired
}

func (s *Session) IsInitialized() bool {
	return s.pinInitialized
}

func (s *Session) IsPairedToCard() bool {
	return s.RemoteCard != nil
}

// Connect opens a secure channel with a card.
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

// Initializes the card with a PIN
// Also creates a secure channel and verifies the PIN that was just set
func (s *Session) Init(pin string) error {
	if s.pinInitialized {
		return ErrAlreadyInitialized
	}

	s.ElementUsageMtex.Lock()
	err := s.cs.Init(pin)
	s.ElementUsageMtex.Unlock()
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

func (s *Session) CreatePhonon() (keyIndex model.PhononKeyIndex, pubkey model.PhononPubKey, err error) {
	if !s.verified() {
		return 0, nil, backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()
	index, pubkey, err := s.cs.CreatePhonon(model.Secp256k1)
	if err == nil {
		s.cache[index] = cachedPhonon{
			pubkeyCached: true,
			infoCached:   true,
			p: &model.Phonon{
				KeyIndex:  index,
				CurveType: model.Secp256k1,
				PubKey:    pubkey,
			},
		}
	}
	return index, pubkey, err
}

func (s *Session) SetDescriptor(p *model.Phonon) error {
	if !s.verified() {
		return backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()
	err := s.cs.SetDescriptor(p)
	if err == nil {
		s.addInfoToCache(p)
	}
	return err
}

func (s *Session) ListPhonons(currencyType model.CurrencyType, lessThanValue uint64, greaterThanValue uint64) ([]*model.Phonon, error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}
	if s.cachePopulated {
		ret := []*model.Phonon{}
		for _, p := range s.cache {
			ret = append(ret, p.p)
		}
		return ret, nil
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	phonons, err := s.cs.ListPhonons(currencyType, lessThanValue, greaterThanValue, false)
	// add listed phonons to the cache
	for _, phonon := range phonons {
		s.addInfoToCache(phonon)
	}

	if currencyType == 0 && lessThanValue == 0 && greaterThanValue == 0 {
		//all phonons were listed, therefore each one can be accounted for in the cache
		s.cachePopulated = true
	}
	return phonons, err
}

func (s *Session) GetPhononPubKey(keyIndex model.PhononKeyIndex, crv model.CurveType) (pubkey model.PhononPubKey, err error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	k, err := s.cs.GetPhononPubKey(keyIndex, crv)
	if err == nil {
		s.addPubKeyToCache(keyIndex, k)
	}
	return k, err
}

func (s *Session) DestroyPhonon(keyIndex model.PhononKeyIndex) (privKey *ecdsa.PrivateKey, err error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	privKey, err = s.cs.DestroyPhonon(keyIndex)
	if err == nil {
		delete(s.cache, keyIndex)
	}
	return privKey, err
}

func (s *Session) IdentifyCard(nonce []byte) (cardPubKey *ecdsa.PublicKey, cardSig *util.ECDSASignature, err error) {
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.IdentifyCard(nonce)
}

func (s *Session) InitCardPairing(receiverCert cert.CardCertificate) ([]byte, error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.InitCardPairing(receiverCert)
}

func (s *Session) CardPair(initPairingData []byte) ([]byte, error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	return s.cs.CardPair(initPairingData)
}

func (s *Session) CardPair2(cardPairData []byte) (cardPair2Data []byte, err error) {
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
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
		return backend.ErrPINNotEntered
	}
	s.ElementUsageMtex.Lock()
	defer s.ElementUsageMtex.Unlock()

	err := s.cs.FinalizeCardPair(cardPair2Data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) SendPhonons(keyIndices []model.PhononKeyIndex) error {
	log.Debug("Sending phonons")
	if !s.verified() && s.RemoteCard != nil {
		return ErrCardNotPairedToCard
	}
	log.Debug("verifying pairing")
	err := s.RemoteCard.VerifyPaired()
	if err != nil {
		return err
	}
	log.Debug("locking mutex")
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
	fmt.Println("unlockingMutex")
	for _, index := range keyIndices {
		delete(s.cache, index)
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
	//invalidate the cache now that new phonons have been received
	s.cachePopulated = false
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
	remConn, err := remote.Connect(s.remoteMessageChan, fmt.Sprintf("https://%s/phonon", u.Host), true)
	if err != nil {
		return fmt.Errorf("unable to connect to remote session: %s", err.Error())
	}
	s.RemoteCard = remConn
	return nil
}

func (s *Session) RemoteConnectionStatus() model.RemotePairingStatus {
	if s.RemoteCard == nil {
		return model.StatusUnconnected
	}
	return s.RemoteCard.PairingStatus()
}

func (s *Session) ConnectToLocalProvider() error {
	lcp := &localCounterParty{
		localSession:  s,
		pairingStatus: model.StatusConnectedToBridge,
	}
	s.RemoteCard = lcp
	connectedCardsAndLCPSessions[s] = lcp
	return nil
}

func (s *Session) ConnectToCounterparty(cardID string) error {
	err := s.RemoteCard.ConnectToCard(cardID)
	if err != nil {
		log.Info("returning error from ConnectRemoteSession")
		return err
	}
	_, err = util.ParseECCPubKey(s.Cert.PubKey)
	if err != nil {
		//we shouldn't get this far and still receive this error
		return err
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

/*
InitDepositPhonons takes a currencyType and a map of denominations to quantity,
Creates the required phonons, deposits them using the configured service for the asset
and upon success sets their descriptors
*/
func (s *Session) InitDepositPhonons(currencyType model.CurrencyType, denoms []*model.Denomination) (phonons []*model.Phonon, err error) {
	log.Debugf("running InitDepositPhonons with data: %v, %v\n", currencyType, denoms)
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
	}
	for _, denom := range denoms {
		p := &model.Phonon{}
		p.KeyIndex, p.PubKey, err = s.CreatePhonon()
		log.Debug("ran CreatePhonons in InitDepositLoop")
		if err != nil {
			log.Error("failed to create phonon for deposit: ", err)
			return nil, err
		}
		p.Denomination = *denom
		p.CurrencyType = currencyType
		p.Address, err = s.chainSrv.DeriveAddress(p)
		if err != nil {
			log.Error("failed to derive address for phonon deposit: ", err)
			return nil, err
		}

		phonons = append(phonons, p)
	}
	return phonons, nil
}

// Phonon Deposit and Redeem higher level methods
type DepositConfirmation struct {
	Phonon           *model.Phonon
	ConfirmedOnChain bool
	ConfirmedOnCard  bool
}

func (s *Session) FinalizeDepositPhonons(confirmations []DepositConfirmation) ([]DepositConfirmation, error) {
	log.Debug("running finalizeDepositPhonon")
	if !s.verified() {
		return nil, backend.ErrPINNotEntered
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
		resp.Name = s.GetCardId()
		resp.Err = nil
		req.Ret <- resp
	case "RequestPairWithRemote":
		req, ok := r.(*model.RequestPairWithRemote)
		if !ok {
			panic("this shouldn't happen.")
		}
		var resp model.ResponsePairWithRemote
		resp.Err = s.PairWithRemoteCard(req.Card)
		log.Debug("Returning pairing stuff")
		req.Ret <- resp
		log.Debug("Done returning pairing stuff")
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

/*
RedeemPhonon takes a phonon and a redemptionAddress as an asset specific address string (usually hex encoded)
and submits a transaction to the asset's chain in order to transfer it to another address
In case the on chain transfer fails, returns the private key as a fallback so that access to the asset is not lost
*/
func (s *Session) RedeemPhonon(p *model.Phonon, redeemAddress string) (transactionData string, privKeyString string, err error) {
	err = s.chainSrv.CheckRedeemable(p, redeemAddress)
	if err != nil {
		return "", "", err
	}

	//Retrieve phonon private key.
	privKey, err := s.DestroyPhonon(p.KeyIndex)
	if err != nil {
		return "", "", err
	}
	privKeyString = util.ECCPrivKeyToHex(privKey)
	transactionData, err = s.chainSrv.RedeemPhonon(p, privKey, redeemAddress)
	if err != nil {
		return "", privKeyString, err
	}

	return transactionData, privKeyString, nil
}

/*
addPubKeyToCache adds the pubkey to an already cached phonon. This is done differently from the addInfoToCache because there are only two
instances where we add information to a preexisting cached phonon, and they need to be handled differently without going through the trouble
of making a fully generic updater that handles all fields
*/
func (s *Session) addPubKeyToCache(i model.PhononKeyIndex, k model.PhononPubKey) {
	cached, ok := s.cache[i]
	// if it didn't already exist, create it
	if !ok {
		s.cache[i] = cachedPhonon{
			p: &model.Phonon{
				KeyIndex: i,
				PubKey:   k,
			},
			pubkeyCached: true,
		}
		// otherwise, add the new descriptor things
	} else {
		cached.p.PubKey = k
		cached.pubkeyCached = true
	}
}

/*addInfoToCache handles merging information we know and information we don't know into the cache*/
func (s *Session) addInfoToCache(p *model.Phonon) {
	cached, ok := s.cache[p.KeyIndex]
	// if it didn't already exist, create it
	if !ok {
		s.cache[p.KeyIndex] = cachedPhonon{
			p:          p,
			infoCached: true,
		}
		// otherwise, add the new descriptor things
	} else {
		cachedPubKey := cached.p.PubKey
		s.cache[p.KeyIndex] = cachedPhonon{
			p:            p,
			pubkeyCached: cached.pubkeyCached,
			infoCached:   true,
		}
		s.cache[p.KeyIndex].p.PubKey = cachedPubKey
	}

}
