package repl

import (
	"strconv"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"
	ishell "github.com/abiosoft/ishell/v2"
)

func createPhonon(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	keyIndex, pubKey, err := activeCard.CreatePhonon()
	if err != nil {
		c.Println("error creating phonon: ", err)
		return
	}
	c.Println("created phonon")
	c.Println("Key Index: ", keyIndex)
	c.Println("Public Key: ", util.ECDSAPubKeyToHexString(pubKey))
}

func listPhonons(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	var currencyType model.CurrencyType = 0
	var lessThanValue float32 = 0
	var greaterThanValue float32 = 0
	var numCorrectArgs = 3

	if len(c.Args) == numCorrectArgs {
		currencyTypeInt, err := strconv.ParseInt(c.Args[0], 10, 0)
		if err != nil {
			c.Println("error parsing currencyType: ", err)
			return
		}
		currencyType = model.CurrencyType(currencyTypeInt)

		lessThanValueRaw, err := strconv.ParseFloat(c.Args[1], 32)
		if err != nil {
			c.Println("error parsing lessThanValue: ", err)
			return
		}
		lessThanValue = float32(lessThanValueRaw)

		greaterThanValueRaw, err := strconv.ParseFloat(c.Args[2], 32)
		if err != nil {
			c.Println("error parsing greaterThanValue: ", err)
			return
		}
		greaterThanValue = float32(greaterThanValueRaw)
	}
	phonons, err := activeCard.ListPhonons(currencyType, lessThanValue, greaterThanValue)
	if err != nil {
		c.Println("error listing phonons: ", err)
		return
	}
	for _, p := range phonons {
		p.PubKey, err = activeCard.GetPhononPubKey(p.KeyIndex)
		c.Println("retrieved pubKey: ", p.PubKey)
		if err != nil {
			c.Printf("error retrieving phonon pubKey at keyIndex %v. err: %v\n", p.KeyIndex, err)
		}
	}
	c.Println("phonons: ")
	for _, p := range phonons {
		c.Printf("%v\n", p)
	}
}

func setDescriptor(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	numCorrectArgs := 3
	if len(c.Args) != numCorrectArgs {
		c.Println("setDescriptor requires %v args", numCorrectArgs)
		return
	}

	keyIndex, err := strconv.ParseUint(c.Args[0], 10, 16)
	if err != nil {
		c.Println("keyIndex could not be parsed: ", err)
		return
	}
	//TODO: Present these options better
	currencyTypeInt, err := strconv.ParseInt(c.Args[1], 10, 0)
	if err != nil {
		c.Println("currencyType could not be parse: ", err)
		return
	}
	currencyType := model.CurrencyType(currencyTypeInt)

	value, err := strconv.ParseFloat(c.Args[2], 32)
	if err != nil {
		c.Println("value could not be parse: ", err)
		return
	}
	c.Println("setting descriptor with values: ", uint16(keyIndex), currencyType, float32(value))
	err = activeCard.SetDescriptor(uint16(keyIndex), currencyType, float32(value))
	if err != nil {
		c.Println("could not set descriptor: ", err)
		return
	}
	c.Println("descriptor set successfully")
	//TODO: wizard?
	//TODO: Resolve SetDescriptor issue on card
}
