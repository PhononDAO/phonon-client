package main

import (
	"encoding/hex"
	"fmt"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/validator"
	"github.com/btcsuite/btcd/btcec"
)

func main() {
	pubkeys := []string{
		"04500369f4a8090ca7a1265854de8ab6019168ffe41d537392e6027c74fffcc558e385686cabd665a7954f5edba6af66d5bb9a7d9dc3602c9b4a5078ce0bf11cd4",
	}
	for _, addresspubkey := range pubkeys {
		h, err := hex.DecodeString(addresspubkey)
		if err != nil {
			fmt.Println("rip")
			fmt.Println(err.Error())
			panic(1)
		}
		k, err := btcec.ParsePubKey(h, btcec.S256())
		if err != nil {
			fmt.Println("fuck")
			fmt.Println(err.Error())
			panic(1)
		}
		fmt.Println(k)
		bcoinClient := validator.NewClient("https://bcoin.gridpl.us", "")
		val := validator.NewBTCValidator(bcoinClient)
		testPhonon := model.Phonon{
			KeyIndex:     0,
			PubKey:       k.ToECDSA(),
			Value:        0,
			CurrencyType: 0,
		}
		res, err := val.Validate(&testPhonon)
		if err != nil {
			fmt.Println("heckin' crap")
			fmt.Println(err.Error())
		}
		fmt.Println(res)
	}
}
