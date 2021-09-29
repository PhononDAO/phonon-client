package repl

import (
	"fmt"
	ishell "github.com/abiosoft/ishell/v2"
)

func initCard(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	c.Println("plase enter new numeric 6 digit PIN")
	pin := c.ReadPassword()
	err := activeCard.Init(pin)
	if err != nil {
		c.Println("unable to initialize card with PIN: ", err)
		return
	}
	updatePrompt()
}

func unlockCard(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	var pin string
	// err declared to avoid double declaration of session index in if block
	var err error

	c.Println("Please enter pin")
	pin = c.ReadPassword()
	//TODO: update chain of functions to return triesRemaining to callers
	err = activeCard.VerifyPIN(pin)
	if err != nil {
		c.Err(fmt.Errorf("Unable to unlock card %s", err.Error()))
		return
	}
	c.Println("card successfully unlocked")
	//TODO: maybe build a little helper class for keeping this updated
	updatePrompt()
}

func changeCardPIN(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}
	c.Println("please enter new numeric 6 digit PIN")
	pin := c.ReadPassword()
	err := activeCard.ChangePIN(pin)
	if err != nil {
		c.Println()
	}
}
