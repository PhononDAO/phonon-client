package repl

import (
	"fmt"
	"strconv"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/orchestrator"
	ishell "github.com/abiosoft/ishell/v2"
)

// variables global to package. possibly to be moved to a struct, but not truly necessary as this package is mostly self contained and un-exported.
var t orchestrator.PhononTerminal

// -1 indicates no card selected. otherwise, selected card is card at the index of selectedCard
var (
	selectedCard int = -1
	// listedSessions holds the sessions from the last time the sessions were listed. this is not automatically updated if a card is plugged in in case the new card is given an index between two existing cards. this assumes each session will keep track of which card it is attached to using a unique identifier for the card.
	listedSessions []*card.Session
)

const standardPrompt string = "Phonon Cmd>"

func Start() {
	shell := ishell.New()
	t = orchestrator.PhononTerminal{}
	// get initial state of orchestrator
	t.RefreshSessions()
	shell.Println("Welcome to the phonon command interface")
	shell.SetPrompt(standardPrompt)

	shell.AddCmd(&ishell.Cmd{
		Name: "refresh",
		Func: refresh,
		Help: "refresh the current state of attached cards",
	})
	shell.AddCmd(&ishell.Cmd{
		Name:    "list",
		Aliases: []string{},
		Func:    listCards,
		Help:    "List available phonon cards for usage",
	})
	shell.AddCmd(&ishell.Cmd{
		Name:    "unlock",
		Aliases: []string{},
		Func:    unlock,
		Help:    "Unlock card with pin entered in password prompt. Optional argument for card index",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "listPhonons",
		Func: listPhonons,
		Help: "List phonons on card. Optional argument for card index",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "create",
		Func: createPhonon,
		Help: "Create a phonon on selected phonon card. Optional argument for card index",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Func: setDescriptor,
		Help: "Set the type and value of phonon. If card is unselected, first argument is index of card containing the phonon. If card is selected, leave it out. Second argument is index of phonon to be descriptor set, third argument is the type of asset to be associated with the phonon, fourth argument is the value of the asset.",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "balance",
		Func: getBalance,
		Help: "Retrieve the type and balance of a phonon on card. First argument is index of the card containing the phonon, and not needed if a card is selected. Second argument is the index of the phonon you wish to see the balance of",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "connect",
		Func: connectRemoteSession,
		Help: "Connect to a remote session",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "receive",
		Func: setReceiveMode,
		Help: "set card at index to receive phonons",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "redeem",
		Func: redeemPhonon,
		Help: "Destroy the phonon at index on card at index and retrieve the priate key (NOTE: THIS WILL DESTROY THE PHONON ON THE CARD. DO NOT RUN THIS WITHOUT BEING READY TO COPY OUT THE PRIVATE KEY",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "card",
		Func: selectCard,
		Help: "Select a card and enter the prompt for the specific card",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "unselect",
		Func: unselectCard,
		Help: "Deselect a card if one is selected",
	})
	shell.Run()
}

func refresh(c *ishell.Context) {
	t.RefreshSessions()
	c.Printf("Sessions Refreshed")
}

func listCards(c *ishell.Context) {
	c.Printf("Sessions: %%v", t.ListSessions())
}

func selectCard(c *ishell.Context) {
	cardIndex, err := getSession(c, 0)
	if err != nil {
		c.Err(fmt.Errorf("no card selected for operation: %s", err.Error()))
		return
	}
	cardShell(c, cardIndex)
}

// getSessionStateAware returns the session if there is a card selected, and otherwise determines the card in use by calling the getSession function
func getSessionStateAware(c *ishell.Context, numArgsNoSession int) (int, error) {
	if selectedCard != -1 {
		return selectedCard, nil
	} else {
		return (getSession(c, numArgsNoSession))
	}
}

func getSession(c *ishell.Context, numArgsNoSession int) (int, error) {
	var cardIndex int
	// error declared here because I don't want to double declare cardIndex in the else clause
	var err error
	if len(listedSessions) == 0 {
		return -1, fmt.Errorf("No cards detected on machine. Possibly try refreshing your session")
	} else if len(listedSessions) == 1 {
		cardIndex = 0
	} else if len(c.Args) == numArgsNoSession {
		var availableCards []string
		for i := range listedSessions {
			// todo: add some sort of naming for cards to tell the difference between them
			availableCards = append(availableCards, strconv.Itoa(i))
		}
		cardIndex = c.MultiChoice(availableCards, "Please select card from list")

	} else if len(c.Args) == numArgsNoSession+1 {
		cardIndex, err = strconv.Atoi(c.Args[0])
		if err != nil {
			return -1, err
		}

	} else {
		return -1, fmt.Errorf("Wrong number of arguments in command line expression")
	}
	return cardIndex, nil
}

func unselectCard(c *ishell.Context) {
	c.SetPrompt(standardPrompt)
	selectedCard = -1
}

func cardShell(c *ishell.Context, index int) {
	if len(listedSessions) < index || index == 0 {
		c.Err(fmt.Errorf("No card found at index %d", index))
		return
	}
	selectedCard = index
	c.SetPrompt(fmt.Sprintf("Card %d >", index))
}

func unlock(c *ishell.Context) {
	var pin string
	// err declared to avoid double declaration of session index in if block
	var err error
	sessionIndex, err := getSessionStateAware(c, 0)
	if err != nil {
		c.Err(err)
		return
	}
	c.Printf("Please enter pin for card: %d", sessionIndex)
	pin = c.ReadPassword()
	err = t.UnlockCard(sessionIndex, pin)
	if err != nil {
		c.Err(fmt.Errorf("Unable to unlock card %d: %s", sessionIndex, err.Error()))
		return
	}
}

func listPhonons(c *ishell.Context) {
	sessionIndex, err := getSession(c, 0)
	if err != nil {
		c.Err(err)
	}
	phonons, err := t.ListPhonons(sessionIndex)
	if err != nil {
		c.Err(fmt.Errorf("Unable to list phonons on card %d: %s", sessionIndex, err.Error()))
		return
	}
	c.Printf("Phonons on card %d: %+v", phonons)
}

func createPhonon(c *ishell.Context) {
	sessionIndex, err := getSession(c, 0)
	if err != nil {
		c.Err(err)
		return
	}
	phononIndex, err := t.CreatePhonon(sessionIndex)
	if err != nil {
		c.Err(fmt.Errorf("Unable to create phonon on card %d: %s", sessionIndex, err.Error()))
		return
	}
	c.Printf("Phonon created on card %d at index %d", sessionIndex, phononIndex)
}

func setDescriptor(c *ishell.Context) {
	sessionIndex, err := getSession(c, 1)
	if err != nil {
		c.Err(err)
		return
	}
	// last argument is index of phonon to use
	phononIndex, err := strconv.Atoi(c.Args[len(c.Args)-1])
	if err != nil {
		c.Err(fmt.Errorf("Unable to parse phonon index %s", err.Error()))
		return
	}
	t.SetDescriptor(sessionIndex, phononIndex, struct{}{})
}

func getBalance(c *ishell.Context) {
	sessionIndex, err := getSession(c, 1)
	if err != nil {
		c.Err(err)
		return
	}
	// last argument is index of phonon to use
	phononIndex, err := strconv.Atoi(c.Args[len(c.Args)-1])
	if err != nil {
		c.Err(fmt.Errorf("Unable to parse phonon index %s", err.Error()))
		return
	}
	c.Printf("balance of card %d is %v", sessionIndex, t.GetBalance(sessionIndex, phononIndex))
}

// todo: this
func connectRemoteSession(c *ishell.Context) {
	var sessionIndex int
	t.ConnectRemoteSession(sessionIndex, struct{}{})
}

// todo: this
func setReceiveMode(c *ishell.Context) {
	var sessionIndex int
	t.SetReceiveMode(sessionIndex)
}

func redeemPhonon(c *ishell.Context) {
	sessionIndex, err := getSession(c, 1)
	if err != nil {
		c.Err(err)
		return
	}
	// last argument is index of phonon to use
	phononIndex, err := strconv.Atoi(c.Args[len(c.Args)-1])
	if err != nil {
		c.Err(fmt.Errorf("Unable to parse phonon index %s", err.Error()))
		return
	}
	c.Printf("Phonon %d on card %d deleted. PublicKey: %v", phononIndex, sessionIndex, t.RedeemPhonon(sessionIndex, phononIndex))
}
