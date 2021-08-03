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
		"03110c89d71731d603059f919e1670cd335cb915bb7a27b56a667ee057a2e78f3e",
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
