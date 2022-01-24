package gui

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/GridPlus/phonon-client/session"
	"github.com/GridPlus/phonon-client/util"
	"github.com/gorilla/mux"
)

//go:embed swagger.yaml
var swaggeryaml []byte

//go:embed swagger
var swagger embed.FS

var t orchestrator.PhononTerminal

func Server() {
	t.RefreshSessions()
	r := mux.NewRouter()
	// sessions
	r.HandleFunc("/genMock", generatemock)
	r.HandleFunc("/listSessions", listSessions)
	r.HandleFunc("/cards/{sessionID}/unlock", unlock)
	r.HandleFunc("/cards/{sessionID}/Pair", pair)
	// phonons
	r.HandleFunc("/cards/{sessionID}/listPhonons", listPhonons)
	r.HandleFunc("/cards/{sessionID}/phonon/{PhononIndex}/setDescriptor", setDescriptor)
	r.HandleFunc("/cards/{sessionID}/phonon/{PhononIndex}/send", send)
	r.HandleFunc("/cards/{sessionID}/phonon/create", createPhonon)
	r.HandleFunc("/cards/{sessionID}/phonon/{PhononIndex}/redeem", redeemPhonon)
	// api docs
	http.Handle("/swagger/", http.FileServer(http.FS(swagger)))
	r.HandleFunc("/swagger.json", serveapi)

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func createPhonon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	index, pubkey, err := sess.CreatePhonon()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pub := util.ECCPubKeyToHexString(pubkey)
	enc := json.NewEncoder(w)
	enc.Encode(struct {
		Index  uint16 `json:"index"`
		PubKey string `json:"pubkey"`
	}{Index: index,
		PubKey: pub})
}

func serveapi(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "swagger.json", time.Time{}, bytes.NewReader(swaggeryaml))
}

func listSessions(w http.ResponseWriter, r *http.Request) {
	sessions := t.ListSessions()

	names := []string{}
	if len(sessions) == 0 {
		http.Error(w, "no cards found", http.StatusNotFound)
		return
	}
	for _, v := range sessions {
		names = append(names, v.GetName())
	}
	enc := json.NewEncoder(w)
	enc.Encode(struct {
		Sessions []string
	}{Sessions: names})
}

func unlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
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

func pair(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
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
		URL string `json:"url"`
	}{}
	err = json.Unmarshal(body, &pairReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = t.ConnectRemoteSession(sess, pairReq.URL)
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

func listPhonons(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	phonons, err := sess.ListPhonons(0, 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, p := range phonons {
		p.PubKey, err = sess.GetPhononPubKey(p.KeyIndex)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(phonons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func setDescriptor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
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

	den, err := model.NewDenomination(inputs.Value)
	if err != nil {
		http.Error(w, "Unable to convert value to base and exponent form for phonon storage: "+err.Error(), http.StatusBadRequest)
	}

	p := &model.Phonon{
		KeyIndex:     uint16(index),
		Denomination: den,
		CurrencyType: model.CurrencyType(inputs.CurrencyType),
	}
	p.KeyIndex = uint16(index)

	err = sess.SetDescriptor(p)
	if err != nil {
		http.Error(w, "Unable to set descriptor", http.StatusBadRequest)
		return
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
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
	err = sess.SendPhonons([]uint16{uint16(index)})
	if err != nil {
		http.Error(w, "unable to send phonons: "+err.Error(), http.StatusInternalServerError)
	}

}

func redeemPhonon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sess, err := selectSession(vars)
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
	privkey, err := sess.DestroyPhonon(uint16(index))
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

func generatemock(w http.ResponseWriter, r *http.Request) {
	err := t.GenerateMock()
	if err != nil {
		http.Error(w, "unable to generate mock", http.StatusInternalServerError)
	}
}

func selectSession(p map[string]string) (*session.Session, error) {
	sessionName, ok := p["sessionID"]
	if !ok {
		fmt.Println("unable to find session")
		return nil, fmt.Errorf("Unable to find sesion")
	}
	sessions := t.ListSessions()
	var targetSession *session.Session
	for _, session := range sessions {
		if session.GetName() == sessionName {
			targetSession = session
			break
		}
	}
	if targetSession == nil {
		return nil, fmt.Errorf("Unable to find sesion")
	}
	return targetSession, nil
}
