package model

import (
	"crypto/ecdsa"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/util"
)

type SessionRequest interface {
	GetName() string
}

type RequestCertificate struct {
	Ret chan ResponseCertificate
}

func (*RequestCertificate) GetName() string {
	return "RequestCertificate"
}

type ResponseCertificate struct {
	Err     error
	Payload *cert.CardCertificate
}

type RequestIdentifyCard struct {
	Ret   chan ResponseIdentifyCard
	Nonce []byte
}

func (*RequestIdentifyCard) GetName() string {
	return "RequestIdentifyCard"
}

type ResponseIdentifyCard struct {
	PubKey *ecdsa.PublicKey
	Sig    *util.ECDSASignature
	Err    error
}

type RequestCardPair1 struct {
	Ret     chan ResponseCardPair1
	Payload []byte
}

func (*RequestCardPair1) GetName() string {
	return "RequestCardPair1"
}

type ResponseCardPair1 struct {
	Err     error
	Payload []byte
}

type RequestFinalizeCardPair struct {
	Ret     chan ResponseFinalizeCardPair
	Payload []byte
}

func (*RequestFinalizeCardPair) GetName() string {
	return "RequestFinalizeCardPair"
}

type ResponseFinalizeCardPair struct {
	Err error
}

type RequestSetRemote struct {
	Ret  chan ResponseSetRemote
	Card CounterpartyPhononCard
}

func (*RequestSetRemote) GetName() string {
	return "RequestSetRemote"
}

type ResponseSetRemote struct {
	Err error
}

type RequestReceivePhonons struct {
	Ret     chan ResponseReceivePhonons
	Payload []byte
}

func (*RequestReceivePhonons) GetName() string {
	return "RequestReceivePhonons"
}

type ResponseReceivePhonons struct {
	Err error
}

type RequestGetName struct {
	Ret chan ResponseGetName
}

func (*RequestGetName) GetName() string {
	return "RequestGetName"
}

type ResponseGetName struct {
	Err  error
	Name string
}

type RequestPairWithRemote struct {
	Ret  chan ResponsePairWithRemote
	Card CounterpartyPhononCard
}

func (*RequestPairWithRemote) GetName() string {
	return "RequestPairWithRemote"
}

type ResponsePairWithRemote struct {
	Err error
}

type RequestSetPaired struct {
	Ret    chan ResponseSetPaired
	Status bool
}

func (*RequestSetPaired) GetName() string {
	return "RequestSetPaired"
}

type ResponseSetPaired struct {
	Err error
}
