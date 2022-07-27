package repl

import (
	"strconv"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/abiosoft/ishell/v2"
)

func cardPairLocal(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	c.Println("starting local card pairing")
	sessions := t.ListSessions()
	var otherCards []*orchestrator.Session
	var otherCardNames []string
	for _, session := range sessions {
		if session != activeCard {
			otherCards = append(otherCards, session)
			otherCardNames = append(otherCardNames, session.GetCardId())
		}
	}
	if len(otherCards) == 0 {
		c.Println("no available cards for pairing found")
		return
	}
	selection := c.MultiChoice(otherCardNames, "please select another card to pair with")
	if selection == -1 {
		c.Println("no card selected. exiting pairing...")
		return
	}
	pairingCard := otherCards[selection]
	c.Println("starting pairing with ", pairingCard.GetCardId())
	err := activeCard.ConnectToLocalProvider()
	if err != nil {
		c.Printf("Error occured in pairing: %s", err.Error())
		return
	}
	err = activeCard.ConnectToCounterparty(otherCards[selection].GetCardId())
	if err != nil {
		c.Printf("Error occured in pairing process: %s", err.Error())
		return
	}
	c.Println("cards successfully paired")
}

func sendPhonons(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	if paired := checkCardPaired(c); !paired {
		return
	}
	var keyIndices []model.PhononKeyIndex
	for _, i := range c.Args {
		keyIndex, err := strconv.ParseUint(i, 10, 16)
		if err != nil {
			c.Println("error parsing arg: ", i)
			c.Println("aborting send operation...")
			return
		}
		keyIndices = append(keyIndices, model.PhononKeyIndex(keyIndex))
	}

	err := activeCard.SendPhonons(keyIndices)
	if err != nil {
		c.Println("error during phonon send: ", err)
		return
	}
}
