package chain

import (
	"crypto/ecdsa"
	"errors"

	"github.com/GridPlus/phonon-client/model"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var ErrMissingPubKey = errors.New("phonon missing pubKey")
var ErrMissingKeyIndex = errors.New("phonon missing KeyIndex")
var ErrUnknownCurrencyType = errors.New("unknown currency type")

type ChainService interface {
	DeriveAddress(p *model.Phonon) (address string, err error)
}

/*Check the phonon's currency type and public key and returns a chain specific
address as a hexstring*/
func DeriveAddress(p *model.Phonon) (address string, err error) {
	switch p.CurrencyType {
	case model.Ethereum:
		//TODO initialize this elsewhere
		eth := &ETHChainService{}
		return eth.DeriveAddress(p.PubKey)
	default:
		return "", ErrUnknownCurrencyType
	}
}

type ETHChainService struct{}

//Derives an ETH address from a phonon's ECDSA Public Key
func (eth *ETHChainService) DeriveAddress(pubKey *ecdsa.PublicKey) (address string, err error) {
	return ethcrypto.PubkeyToAddress(*pubKey).Hex(), nil
}
