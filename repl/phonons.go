package repl

import (
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
	//TODO:
	// sessionIndex, err := getSession(c, 0)
	// if err != nil {
	// 	c.Err(err)
	// }
	// phonons, err := activeCard.ListPhonons()
	// phonons, err := t.ListPhonons(sessionIndex)
	// if err != nil {
	// 	c.Err(fmt.Errorf("Unable to list phonons on card %d: %s", sessionIndex, err.Error()))
	// 	return
	// }
	// c.Printf("Phonons on card %d: %+v", phonons)
}

func setDescriptor(c *ishell.Context) {
	// sessionIndex, err := getSession(c, 1)
	// if err != nil {
	// 	c.Err(err)
	// 	return
	// }
	// // last argument is index of phonon to use
	// phononIndex, err := strconv.Atoi(c.Args[len(c.Args)-1])
	// if err != nil {
	// 	c.Err(fmt.Errorf("Unable to parse phonon index %s", err.Error()))
	// 	return
	// }
	// t.SetDescriptor(sessionIndex, phononIndex, struct{}{})
}
