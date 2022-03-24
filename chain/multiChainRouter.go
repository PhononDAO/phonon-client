package chain

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/phonon-client/model"
)

/*MultiChainRouter is a ChainService supporting multiple currencyTypes by
encapsulating a specific hardcoded collection of ChainServices keyed by CurrencyType*/
type MultiChainRouter struct {
	chainServices map[model.CurrencyType]ChainService
}

var ErrCurrencyTypeUnsupported = errors.New("currency type is not supported")

//Initialize all supported chain services at start
func NewMultiChainRouter() (*MultiChainRouter, error) {
	mcr := &MultiChainRouter{
		chainServices: make(map[model.CurrencyType]ChainService),
	}
	err := mcr.initBuiltinChainServices()
	if err != nil {
		return nil, err
	}
	return mcr, nil
}

func (mcr *MultiChainRouter) initBuiltinChainServices() (err error) {
	mcr.chainServices[model.Ethereum], err = NewEthChainService()
	if err != nil {
		return err
	}
	return nil
}

//ChainService interface methods
func (mcr *MultiChainRouter) DeriveAddress(p *model.Phonon) (address string, err error) {
	chain, ok := mcr.chainServices[p.CurrencyType]
	if !ok {
		return "", ErrCurrencyTypeUnsupported
	}
	return chain.DeriveAddress(p)
}

func (mcr *MultiChainRouter) RedeemPhonon(p *model.Phonon, privKey *ecdsa.PrivateKey, redeemAddress string) (transactionData string, err error) {
	chain, ok := mcr.chainServices[p.CurrencyType]
	if !ok {
		return "", ErrCurrencyTypeUnsupported
	}
	return chain.RedeemPhonon(p, privKey, redeemAddress)
}
