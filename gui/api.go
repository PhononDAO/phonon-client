package gui

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

//go:embed frontend/build/*
var frontend embed.FS

//go:embed swagger.yaml
var swaggeryaml []byte

//go:embed swagger
var swagger embed.FS

type apiSession struct {
	t *orchestrator.PhononTerminal
}

func Server(port string, certFile string, keyFile string, mock bool) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.Debug("starting local api server")
	session := apiSession{orchestrator.NewPhononTerminal()}

	//initialize cache map
	if mock {
		//Start server with a mock and ignore actual cards
		_, err := session.t.GenerateMock()
		log.Debug("mock generated")
		if err != nil {
			log.Error("unable to generate mock during REST server startup: ", err)
			return
		}
	} else {
		_, err := session.t.RefreshSessions()
		if err != nil {
			log.Error("unable to refresh card sessions during REST server startup: ", err)
		}
	}
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Origin"},
		AllowCredentials: true,
	})
	handler := c.Handler(r)
	// sessions
	r.HandleFunc("/genMock", session.generatemock)
	r.HandleFunc("/listSessions", session.listSessions)
	r.HandleFunc("/cards/{sessionID}/init", session.init)
	r.HandleFunc("/cards/{sessionID}/unlock", session.unlock)
	r.HandleFunc("/cards/{sessionID}/pair", session.pair)
	r.HandleFunc("/cards/{sessionID}/name", session.setName)
	// phonons
	r.HandleFunc("/cards/{sessionID}/listPhonons", session.listPhonons)
	r.HandleFunc("/cards/{sessionID}/phonon/{PhononIndex}/setDescriptor", session.setDescriptor)
	r.HandleFunc("/cards/{sessionID}/phonon/send", session.send)
	r.HandleFunc("/cards/{sessionID}/phonon/create", session.createPhonon)
	r.HandleFunc("/cards/{sessionID}/phonon/redeem", session.redeemPhonons)
	r.HandleFunc("/cards/{sessionID}/phonon/{PhononIndex}/export", session.exportPhonon)
	r.HandleFunc("/cards/{sessionID}/phonon/initDeposit", session.initDepositPhonons)
	r.HandleFunc("/cards/{sessionID}/phonon/finalizeDeposit", session.finalizeDepositPhonons)
	r.HandleFunc("/cards/{sessionID}/connect", session.ConnectRemote)
	r.HandleFunc("/cards/{sessionID}/connectionStatus", session.RemoteConnectionStatus)
	r.HandleFunc("/cards/{sessionID}/connectLocal", session.ConnectLocal)
	r.HandleFunc("/checkDenomination", verifyDenomination)
	// api docs
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/", http.FileServer(http.FS(swagger))))
	r.HandleFunc("/swagger.json", serveAPIFunc(port))

	// log sink
	r.HandleFunc("/logs", logsink)
	// frontend
	stripped, err := fs.Sub(frontend, "frontend/build")
	if err != nil {
		log.Fatal("unable to setup filesystem to serve frontend: " + err.Error())
	}
	r.PathPrefix("/").Handler(http.FileServer(http.FS(stripped)))

	http.Handle("/", r)
	log.Debug("listening for incoming connections on " + port)
	fmt.Println("listen and serve")
	go func() {
		if certFile != "" && keyFile != "" {
			err := http.ListenAndServeTLS(":"+port, certFile, keyFile, handler)
			if err != nil {
				log.Fatal("could not start GUI REST server on SSL: ", err)
			}
		} else {
			err := http.ListenAndServe(":"+port, handler)
			if err != nil {
				log.Fatal("could not start GUI REST server", err)
			}
		}
	}()
	// setup channel to end the application
	browser.OpenURL("http://localhost:" + port + "/")
	// start the systray Icon
	SystrayIcon(port)
}

func verifyDenomination(w http.ResponseWriter, r *http.Request) {
	tocheckBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to parse request", http.StatusBadRequest)
		return
	}
	toCheckString := string(tocheckBytes)
	toCheck := new(big.Int)
	toCheck, ok := toCheck.SetString(toCheckString, 10)
	if !ok {
		http.Error(w, "unable to coerce request string to integer", http.StatusBadRequest)
		return
	}
	_, err = model.NewDenomination(toCheck)
	if err != nil {
		http.Error(w, "cannot make phonon denomination from value: "+err.Error(), http.StatusBadRequest)
		return
	}
	// return 200 if it works
}

func logsink(w http.ResponseWriter, r *http.Request) {
	var msg map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Errorf("unable to decode logs from frontend: %s\n", err)
		http.Error(w, "unable to decode logs", http.StatusBadRequest)
		return
	}
	lvl, ok := msg["level"]
	if !ok {
		log.WithFields(log.Fields(msg)).Debug()
		return
	} else {
		lvlInt, err := parseJSLogLevel(lvl)
		if err != nil {
			log.Debug(fmt.Sprintf("unable to decode log level from frontend: %s. Defaulting to debug", err.Error()))
			log.WithFields(log.Fields(msg)).Debug()
		} else {
			switch lvlInt {
			case jsLevelDebug:
				log.WithFields(log.Fields(msg)).Debug()
			case jsLevelInfo:
				log.WithFields(log.Fields(msg)).Info()
			case jsLevelWarn:
				log.WithFields(log.Fields(msg)).Warn()
			case jsLevelError:
				log.WithFields(log.Fields(msg)).Error()
			// no critical with logrus, so using error
			case jsLevelCritical:
				log.WithFields(log.Fields(msg)).Error()
			default:
				log.Debug("unable to decode log level from frontend. Defaulting to debug")
				log.WithFields(log.Fields(msg)).Debug()
			}
		}
	}
}

const (
	jsLevelDebug    = 20
	jsLevelInfo     = 30
	jsLevelWarn     = 40
	jsLevelError    = 50
	jsLevelCritical = 60
)

func parseJSLogLevel(input interface{}) (int, error) {
	var levelMap map[string]interface{}
	if reflect.TypeOf(input) != reflect.TypeOf(map[string]interface{}{}) {
		return 0, fmt.Errorf("unable to parse level data from map")
	} else {
		levelMap = input.(map[string]interface{})
	}
	lvlraw, ok := levelMap["value"]
	if !ok {
		return 0, fmt.Errorf("unable to find value key within level object")
	}
	lvlFloat64, ok := lvlraw.(float64)
	if !ok {
		return 0, fmt.Errorf("unable to parse level value: %v into number", lvlraw)
	}
	lvlInt := int(lvlFloat64)
	return lvlInt, nil
}

func (apiSession apiSession) createPhonon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	index, pubKey, err := sess.CreatePhonon()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(struct {
		Index  model.PhononKeyIndex `json:"index"`
		PubKey string               `json:"pubkey"`
	}{Index: index,
		PubKey: pubKey.String()})
}

func (apiSession *apiSession) initDepositPhonons(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var depositPhononReq struct {
		CurrencyType  model.CurrencyType
		Denominations []*model.Denomination
	}
	err = json.NewDecoder(r.Body).Decode(&depositPhononReq)
	if err != nil {
		log.Error("unable to decode initDeposit request")
		return
	}
	log.Debug("depositPhononReq: ", depositPhononReq)
	log.Debug("denoms: ", depositPhononReq.Denominations)
	phonons, err := sess.InitDepositPhonons(depositPhononReq.CurrencyType, depositPhononReq.Denominations)
	if err != nil {
		log.Error("unable to create phonons for deposit. err: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(phonons)
	if err != nil {
		log.Error("unable to encode outgoing depositPhonons response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) finalizeDepositPhonons(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var depositConfirmations []orchestrator.DepositConfirmation
	err = json.NewDecoder(r.Body).Decode(&depositConfirmations)
	if err != nil {
		log.Error("unable to decode depositConfirmations json. err: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := sess.FinalizeDepositPhonons(depositConfirmations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(ret)
	if err != nil {
		log.Error("unable to encode outgoing deposit confirmation response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) redeemPhonons(w http.ResponseWriter, r *http.Request) {
	sess, err := apiSession.sessionFromMuxVars(mux.Vars(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	type redeemPhononRequest struct {
		P             *model.Phonon
		RedeemAddress string
	}
	var reqs []*redeemPhononRequest
	err = json.NewDecoder(r.Body).Decode(&reqs)
	if err != nil {
		log.Error("unable to decode redeemPhonons json. err: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(reqs) == 0 {
		log.Error("request data empty")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, req := range reqs {
		log.Debugf("received redeem phonon %+v", req.P)
		log.Debug("received redeem address: ", req.RedeemAddress)
	}
	//TODO: Validate data contains what it needs to
	type redeemPhononResp struct {
		TransactionData string
		PrivKey         string
		Err             string
	}
	var resps []*redeemPhononResp
	for _, req := range reqs {
		var respErr string
		var transactionData string
		var privKeyString string
		transactionData, privKeyString, err = sess.RedeemPhonon(req.P, req.RedeemAddress)
		//If err capture the error message as a string, else return string value ""
		if err != nil {
			respErr = err.Error()
		}
		resps = append(resps, &redeemPhononResp{
			TransactionData: transactionData,
			PrivKey:         privKeyString,
			Err:             respErr,
		})
	}

	success := true
	for _, res := range resps {
		if res.Err != "" {
			success = false
		}
	}
	if !success {
		w.WriteHeader(http.StatusInternalServerError)
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(resps)
	if err != nil {
		log.Error("unable to encode outgoing redeem response")
		return
	}
}

func serveapi(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "swagger.json", time.Time{}, bytes.NewReader(swaggeryaml))
}

func serveAPIFunc(port string) func(w http.ResponseWriter, r *http.Request) {
	swaggerTemplateFile := string(swaggeryaml)
	templ, err := template.New("swaggeryaml").Parse(swaggerTemplateFile)
	if err != nil {
		// this shouldn't happen. this is to make sure it fails in testing if it's set up wrong
		log.Fatal("unable to render swagger template. Exiting")
	}
	buff := bytes.NewBuffer([]byte{})
	err = templ.Execute(buff, port)
	if err != nil {
		log.Fatal("Unable to render port into swagger yaml, Exting")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "swagger.json", time.Time{}, bytes.NewReader(buff.Bytes()))
	}
}

func (apiSession apiSession) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions := apiSession.t.ListSessions()
	if len(sessions) == 0 {
		http.Error(w, "no card sessions found", http.StatusNotFound)
		return
	}
	log.Debug("listSessions endpoint found sessions: ", sessions)
	type SessionStatus struct {
		Id             string
		Name           string
		Initialized    bool
		TerminalPaired bool
		PinVerified    bool
	}
	sessionStatuses := make([]*SessionStatus, 0)

	for _, v := range sessions {
		name, err := v.GetName()
		if err != nil {
			log.Error("unable to retrieve friendly name: " + err.Error())
			name = ""
		}
		sessionStatuses = append(sessionStatuses,
			&SessionStatus{
				Id:             v.GetCardId(),
				Name:           name,
				Initialized:    v.IsInitialized(),
				TerminalPaired: v.IsPairedToTerminal(),
				PinVerified:    v.IsUnlocked(),
			})
	}

	log.Debug("listSessions sessionStatuses: ", sessionStatuses)
	enc := json.NewEncoder(w)
	enc.Encode(sessionStatuses)
}

func (apiSession apiSession) init(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if sess.IsInitialized() {
		http.Error(w, "card is already initialized", http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	initReq := struct {
		Pin string
	}{}
	err = json.Unmarshal(body, &initReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = sess.Init(initReq.Pin)
	if err != nil && err.Error() != "bad response 6983: unexpected sw in secure channel" {
		http.Error(w, fmt.Errorf("unable to initialize card with given PIN. err: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	/*Workaround for error in mutual auth that occurs in rest of session after INIT
	  Blows away all sessions, but shouldn't cause problems except when a mock is configured
	  or another card is already unlocked, which should generally not come up when
	  Initializing a real card
	*/
	_, err = apiSession.t.RefreshSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (apiSession apiSession) unlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	unlockReq := struct {
		Pin string
	}{}
	err = json.Unmarshal(body, &unlockReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = sess.VerifyPIN(unlockReq.Pin)
	if err != nil {
		http.Error(w, "Unable to validate pin", http.StatusBadRequest)
		return
	}
}
func (apiSession apiSession) ConnectRemote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	ConnectionReq := struct {
		URL string `json:"url"`
	}{}
	err = json.Unmarshal(body, &ConnectionReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = sess.ConnectToRemoteProvider(ConnectionReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (apiSession apiSession) RemoteConnectionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	//Fetch connection status from session and encode response object
	type connectionStatusResp struct {
		ConnectionStatus model.RemotePairingStatus
	}
	resp := &connectionStatusResp{
		ConnectionStatus: sess.RemoteConnectionStatus(),
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(resp)
	if err != nil {
		log.Error("unable to encode outgoing RemoteConnectionStatus response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) ConnectLocal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = sess.ConnectToLocalProvider()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) pair(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	pairReq := struct {
		CardID string `json:"cardID"`
	}{
		CardID: "",
	}
	err = json.Unmarshal(body, &pairReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = sess.ConnectToCounterparty(pairReq.CardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (apiSession apiSession) setName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	nameReq := struct {
		Name string
	}{}
	err = json.Unmarshal(body, &nameReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = sess.SetName(nameReq.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type phonRet struct {
	Index  int    `json:"index"`
	PubKey string `json:"pubKey"`
	Type   int    `json:"type"`
	Value  int    `json:"value"`
}

func (apiSession apiSession) listPhonons(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var phonons []*model.Phonon
	phonons, err = sess.ListPhonons(0, 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, p := range phonons {
		if p.PubKey == nil {
			p.PubKey, err = sess.GetPhononPubKey(p.KeyIndex, p.CurveType)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(phonons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) setDescriptor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	phononIndex, ok := vars["PhononIndex"]
	if !ok {
		http.Error(w, "Phonon not found", http.StatusNotFound)
		return
	}
	index, err := strconv.ParseUint(phononIndex, 10, 16)
	if err != nil {
		http.Error(w, "Unable to convert index to int:"+err.Error(), http.StatusBadRequest)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}

	inputs := struct {
		CurrencyType int `json:"currencyType"`
		Value        int `json:"value"`
	}{}
	json.Unmarshal(b, &inputs)

	den, err := model.NewDenomination(big.NewInt(int64(inputs.Value)))
	if err != nil {
		http.Error(w, "Unable to convert value to base and exponent form for phonon storage: "+err.Error(), http.StatusBadRequest)
	}

	p := &model.Phonon{
		KeyIndex:     model.PhononKeyIndex(index),
		Denomination: den,
		CurrencyType: model.CurrencyType(inputs.CurrencyType),
	}
	p.KeyIndex = model.PhononKeyIndex(index)
	err = sess.SetDescriptor(p)
	if err != nil {
		http.Error(w, "Unable to set descriptor", http.StatusBadRequest)
		return
	}
}

func (apiSession apiSession) send(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	inputs := []model.Phonon{}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.Unmarshal(bodyBytes, &inputs)
	toSend := []model.PhononKeyIndex{}
	for _, phonon2send := range inputs {
		toSend = append(toSend, phonon2send.KeyIndex)
	}
	err = sess.SendPhonons(toSend)

	if err != nil {
		http.Error(w, "unable to send phonons: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) exportPhonon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := apiSession.sessionFromMuxVars(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	phononIndex, ok := vars["PhononIndex"]
	if !ok {
		http.Error(w, "Phonon not found", http.StatusNotFound)
		return
	}
	index, err := strconv.ParseUint(phononIndex, 10, 16)
	if err != nil {
		http.Error(w, "Unable to convert index to int:"+err.Error(), http.StatusBadRequest)
		return
	}
	privkey, err := sess.DestroyPhonon(model.PhononKeyIndex(index))
	if err != nil {
		http.Error(w, "Unable to redeem phonon: "+err.Error(), http.StatusInternalServerError)
		return
	}
	ret := struct {
		PrivateKey string `json:"privateKey"`
	}{PrivateKey: fmt.Sprintf("%x", privkey.D)}
	enc := json.NewEncoder(w)
	err = enc.Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) generatemock(w http.ResponseWriter, r *http.Request) {
	_, err := apiSession.t.GenerateMock()
	if err != nil {
		http.Error(w, "unable to generate mock", http.StatusInternalServerError)
		return
	}
}

func (apiSession apiSession) sessionFromMuxVars(p map[string]string) (*orchestrator.Session, error) {
	sessionName, ok := p["sessionID"]
	if !ok {
		fmt.Println("unable to find session")
		return nil, fmt.Errorf("unable to find session")
	}
	sessions := apiSession.t.ListSessions()
	var targetSession *orchestrator.Session
	for _, session := range sessions {
		if session.GetCardId() == sessionName {
			targetSession = session
			break
		}
	}
	if targetSession == nil {
		return nil, fmt.Errorf("unable to find session")
	}
	return targetSession, nil
}
