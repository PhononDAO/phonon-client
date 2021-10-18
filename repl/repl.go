package repl

import (
	"fmt"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/orchestrator"
	ishell "github.com/abiosoft/ishell/v2"
)

var (
	t          *orchestrator.PhononTerminal
	shell      *ishell.Shell
	activeCard *card.Session
)

const standardPrompt string = "Phonon>"

func Start() {
	shell = ishell.New()
	t = &orchestrator.PhononTerminal{}
	// get initial state of orchestrator
	shell.Println("Welcome to the phonon command interface")
	shell.SetPrompt(standardPrompt)

	shell.AddCmd(&ishell.Cmd{
		Name: "refresh",
		Func: refresh,
		Help: "Check for attached phonon cards. Restarts all phonon card sessions.",
	})
	shell.AddCmd(&ishell.Cmd{
		Name:    "listCards",
		Aliases: []string{},
		Func:    listCards,
		Help:    "List available phonon cards for usage",
	})
	shell.AddCmd(&ishell.Cmd{
		Name:    "unlock",
		Aliases: []string{},
		Func:    unlockCard,
		Help:    "Unlock card by entering PIN in password prompt.",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "activate",
		Func: activateCard,
		Help: "Activate a specific card",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "deactivate",
		Func: deactivateCard,
		Help: "Deselect a card if one is selected",
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "init",
		Func: initCard,
		Help: "Initialize the active card with a PIN",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "changePin",
		Func: changeCardPIN,
		Help: "Change the active card's PIN",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Func: listPhonons,
		Help: `List phonons on card. Optionally takes arguments to filter by.
		       Args: [CurrencyType] [lessThanValue] [greaterThanValue]`,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "create",
		Func: createPhonon,
		Help: "Create a new phonon key on active card",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "setDescriptor",
		Func: setDescriptor,
		Help: `Set the metadata associated with this phonon.
		       Args: [KeyIndex] [CurrencyType] [Value]`,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "redeem",
		Func: redeemPhonon,
		Help: "Destroy the phonon at index on card at index and retrieve the priate key (NOTE: THIS WILL DESTROY THE PHONON ON THE CARD. DO NOT RUN THIS WITHOUT BEING READY TO COPY OUT THE PRIVATE KEY",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "pairLocal",
		Func: cardPairLocal,
		Help: "Pair with another phonon card to establish a secure connection for the exchange of phonons.",
	})
	
	shell.AddCmd(&ishell.Cmd{
		Name: "sendPhonons",
		Func: sendPhonons,
		Help: "Send phonons to paired card",
	})
	// shell.AddCmd(&ishell.Cmd{
	// 	Name: "balance",
	// 	Func: getBalance,
	// 	Help: "Retrieve the type and balance of a phonon on card. First argument is index of the card containing the phonon, and not needed if a card is selected. Second argument is the index of the phonon you wish to see the balance of",
	// })

	 shell.AddCmd(&ishell.Cmd{
	 	Name: "pairRemote",
	 	Func: connectRemoteSession,
	 	Help: "Connect to a remote server",
	 })

	 shell.AddCmd(&ishell.Cmd{
	 	Name:     "mock",
	 	Func:     addMock,
	 	Help:     "make a mock card and add it to the session",
	 })
	// shell.AddCmd(&ishell.Cmd{
	// 	Name: "receive",
	// 	Func: setReceiveMode,
	// 	Help: "set card at index to receive phonons",
	// })

	//Automatically refresh connections as the user is dropped into the shell
	shell.Process("refresh")
	shell.Run()
}

//internal bookkeeping method to set a card to receive subsequent commands
func setActiveCard(c *ishell.Context, s *card.Session) {
	activeCard = s
	updatePrompt()
	c.Printf("%v selected\n", activeCard.GetName())
}

//Updates the prompt to display the status of the active card
func updatePrompt() {
	if activeCard == nil {
		shell.SetPrompt(standardPrompt)
	}
	cardName := activeCard.GetName()
	var status string
	if !activeCard.IsInitialized() {
		status = "-uninitialized"
	} else if !activeCard.IsUnlocked() {
		status = "-locked"
	} else {
		status = ""
	}
	shell.SetPrompt(cardName + status + ">")
}

//checkActiveCard provides a guard function for shell commands to check that there is a card ready to use before proceeding
//should generally be called at the top of shell functions which require a connected card
func checkActiveCard(c *ishell.Context) bool {
	if activeCard == nil {
		c.Println("please select a card before attempting to unlock")
		return false
	}
	return true
}

//checkCardPaired provides a guard function for shell commands to check that a card is paired with a remote
//before performing an operation where this is required, such as sending or receiving phonons
func checkCardPaired(c *ishell.Context) bool {
	if !activeCard.IsPairedToCard() {
		c.Println("card must be paired with remote card to complete this operation")
		return false
	}
	return true
}

func refresh(c *ishell.Context) {
	c.Println("refreshing sessions")
	sessions, err := t.RefreshSessions()
	if err != nil {
		c.Printf("error refreshing sessions: %v", err)
	}
	if len(sessions) == 0 {
		c.Println("no attached cards detected")
	} else if len(sessions) == 1 {
		c.Println("one attached card detected, setting as active")
		setActiveCard(c, sessions[0])
	} else {
		c.Println("multiple cards detected, please use activate command to choose one")
	}
}

func listCards(c *ishell.Context) {
	sessions := t.ListSessions()
	if len(sessions) == 0 {
		c.Println("no cards found")
	}
	c.Println("available cards: ")
	for _, s := range sessions {
		c.Println(s.GetName())
	}
}

func activateCard(c *ishell.Context) {
	sessions := t.ListSessions()
	var sessionNames []string
	for _, session := range sessions {
		sessionNames = append(sessionNames, session.GetName())
	}

	selection := c.MultiChoice(sessionNames, "please select an available card")
	//MulticChoice() returns -1 if nothing is selected
	if selection == -1 {
		fmt.Println("no card selected")
	} else {
		setActiveCard(c, sessions[selection])
	}
}

func deactivateCard(c *ishell.Context) {
	c.SetPrompt(standardPrompt)
	activeCard = nil
}

// func cardShell(c *ishell.Context, index int) {
// 	if len(listedSessions) < index || index == 0 {
// 		c.Err(fmt.Errorf("No card found at index %d", index))
// 		return
// 	}
// 	activeCard = index
// 	c.SetPrompt(fmt.Sprintf("Card %d >", index))
// }

func getBalance(c *ishell.Context) {
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
	// c.Printf("Balance of card %d is %v", sessionIndex, t.GetBalance(sessionIndex, phononIndex))
}

func connectRemoteSession(c *ishell.Context) {
	fmt.Println("connecting to remote")
	if len(c.Args) != 1{
		fmt.Println("wrong number of arguments given")
		return
	}
	card2Connect2 := c.Args[0]
	err := t.ConnectRemoteSession(activeCard,card2Connect2)
	if err != nil{
		c.Err(err)
	}
}

// todo: this
func setReceiveMode(c *ishell.Context) {
	var sessionIndex int
	t.SetReceiveMode(sessionIndex)
}
