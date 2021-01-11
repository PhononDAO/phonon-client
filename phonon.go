package phonon

import (
	"fmt"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/chain"
	"github.com/GridPlus/phonon-client/model"
	log "github.com/sirupsen/logrus"
)

//High level card interface, intended to encapsulate the terminals view of the distinct
//operations needed to handle phonon operations with a smartcard
type Card interface {
	Select() error

	//Opens a secure channel between the terminal and the card to send further operations
	OpenChannel() (string, error)

	//Open a secure mutually authenticated channel between two cards
	MutualAuthChannel() error
	VerifyPin() error
	ListPhonons(limit int, filterType string, filterValue []byte) (phonons []model.Phonon, err error)

	CreatePhonons(n int) (pubKeys [][]byte, err error)
	SetDescriptors(phonons []model.Phonon) error
	//Build an encrypted transaction to send phonons to another card
	SendPhonons(phononIDs []int) (transaction []byte, err error)

	//is this needed? Depends on if terminals are cooperative or if one terminal handles the entire transaction
	ReceivePhonons(transaction []byte) (err error)

	DestroyPhonon(phononID string) (err error)
}

type Chain interface {
	CreatePhonons(pubKeys [][]byte, denominations model.CoinList) ([]model.Phonon, error)
}

type Session struct {
	sc Card
}

func NewSession() (Session, error) {
	sc := card.MockCard{}
	err := sc.VerifyPin()
	if err != nil {
		return Session{}, err
	}
	err = sc.MutualAuthChannel()
	if err != nil {
		return Session{}, err
	}
	return Session{
		sc: card.MockCard{},
	}, nil
}

func SelectChain(assetID model.CryptoAsset, chainID model.CryptoChain) Chain {
	//TODO: switching logic to select the correct blockchain interface depending on the chain and asset type
	c := new(chain.MockChain)
	return c
}

func (s *Session) Deposit(assetID model.CryptoAsset, chainID model.CryptoChain, denominations model.CoinList) error {
	chain := SelectChain(assetID, chainID)

	//calculate length from denominations list
	freshPhonons, err := s.sc.CreatePhonons(1)
	if err != nil {
		log.Error("could not create new phonons on card. err: ", err)
		return err
	}

	phonons, err := chain.CreatePhonons(freshPhonons, denominations)
	if err != nil {
		log.Error("could not create phonons on chain. err: ", err)
		return err
	}

	//TODO: retry probably? OR, retry is the responsibility of the implementation
	err = s.sc.SetDescriptors(phonons)
	if err != nil {
		log.Error("unable to set descriptors on card. err: ", err)
		return err
	}
	log.Info("I love phonons: ", phonons)
	//Set descriptor on cards for new phonons
	return nil
}

//TODO implement receiving filters on this
func (s *Session) ListPhonons(limit int) ([]model.Phonon, error) {
	phonons, err := s.sc.ListPhonons(limit, "", nil)
	if err != nil {
		log.Error("unable to list phonons from card. err: ", err)
		return nil, err
	}
	//TODO: break out the print into a 'frontend' terminal output module
	fmt.Printf("%+v", phonons)
	return phonons, nil
}

// func (s *Session) TransferPhonons()
