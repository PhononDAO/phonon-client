package chain

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/phonon-core/pkg/model"
)

var ErrMissingPubKey = errors.New("phonon missing pubKey")
var ErrMissingKeyIndex = errors.New("phonon missing KeyIndex")
var ErrUnknownCurrencyType = errors.New("unknown currency type")

type ChainService interface {
	DeriveAddress(p *model.Phonon) (address string, err error)
	CheckRedeemable(p *model.Phonon, redeemAddress string) (err error)
	RedeemPhonon(p *model.Phonon, privKey *ecdsa.PrivateKey, redeemAddress string) (transactionData string, err error)
}
