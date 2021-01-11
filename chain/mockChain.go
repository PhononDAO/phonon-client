package chain

import (
	"github.com/GridPlus/phonon-client/model"
)

type MockChain struct{}

func (c MockChain) CreatePhonons(pubKeys [][]byte, denominations model.CoinList) ([]model.Phonon, error) {
	phonons := make([]model.Phonon, 0)
	pubKeysIndex := 0
	for k, v := range denominations {
		//Create v phonons for each denomination k
		for i := 0; i < v; i++ {
			phonons = append(phonons, model.Phonon{
				ID: pubKeys[pubKeysIndex],
				//TODO: probably automate TLV creation
				Descriptor: []model.Tlv{
					{
						Tag:   "AssetType",
						Value: byte(model.Test),
					},
					{
						Tag:   "Value",
						Value: byte(k),
					},
				},
			})
			pubKeysIndex++
		}

	}
	return phonons, nil
}
