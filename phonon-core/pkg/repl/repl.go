package repl

import (
	"fmt"
	"strings"

	"github.com/GridPlus/phonon-core/pkg/orchestrator"

	ishell "github.com/abiosoft/ishell/v2"
)

var (
	t          *orchestrator.PhononTerminal
	shell      *ishell.Shell
	activeCard *orchestrator.Session
)

const standardPrompt string = "Phonon>"

func Start() {
	shell = ishell.New()
	t = orchestrator.NewPhononTerminal()
	// get initial state of orchestrator
	shell.Println("Welcome to the phonon command interface")
	shell.SetPrompt(standardPrompt)

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
		       Args: [KeyIndex] [CurrencyType] [ChainID] [Value]`,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "redeem",
		Func: redeemPhonon,
		Help: "Destroy the phonon at index on card at index and retrieve the private key (NOTE: THIS WILL DESTROY THE PHONON ON THE CARD. DO NOT RUN THIS WITHOUT BEING READY TO COPY OUT THE PRIVATE KEY",
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
		Name: "connectRemote",
		Func: connectRemoteSession,
		Help: "Connect to a remote server",
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "connectLocal",
		Func: connectLocalSession,
		Help: "Connect to a local card",
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "pairCounterparty",
		Func: pairWithCounterparty,
		Help: `pair with a counterparty already connected to the same counterparty provider
		       Args: [counterpartyCardID]`,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "mock",
		Func: addMock,
		Help: "make a mock card and add it to the session",
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "setName",
		Func: setName,
		Help: "set the name of the active card",
	})
	//Automatically refresh connections as the user is dropped into the shell
	shell.Process("refresh")
	shell.Run()
}

func getDisplayName(activeCard *orchestrator.Session) string {
	name, err := activeCard.GetName()
	if err != nil || name == "" {
		return activeCard.GetCardId()
	} else {
		return name
	}
}

// internal bookkeeping method to set a card to receive subsequent commands
func setActiveCard(c *ishell.Context, s *orchestrator.Session) {
	activeCard = s
	updatePrompt()
	c.Printf("%v selected\n", getDisplayName(activeCard))
}

// Updates the prompt to display the status of the active card
func updatePrompt() {
	if activeCard == nil {
		shell.SetPrompt(standardPrompt)
	}

	var status string
	if !activeCard.IsInitialized() {
		status = "-uninitialized"
	} else if !activeCard.IsUnlocked() {
		status = "-locked"
	} else {
		status = ""
	}
	shell.SetPrompt(fmt.Sprintf("%v%v>", getDisplayName(activeCard), status))
}

// checkActiveCard provides a guard function for shell commands to check that there is a card ready to use before proceeding
// should generally be called at the top of shell functions which require a connected card
func checkActiveCard(c *ishell.Context) bool {
	if activeCard == nil {
		c.Println("please select a card before attempting to unlock")
		return false
	}
	return true
}

// checkCardPaired provides a guard function for shell commands to check that a card is paired with a remote
// before performing an operation where this is required, such as sending or receiving phonons
func checkCardPaired(c *ishell.Context) bool {
	if !activeCard.IsPairedToCard() {
		c.Println("card must be paired with remote card to complete this operation")
		return false
	}
	return true
}

func listCards(c *ishell.Context) {
	sessions := t.ListSessions()
	if len(sessions) == 0 {
		c.Println("no cards found")
	} else {
		for _, s := range sessions {
			name, err := s.GetName()
			if err != nil || name == "" {
				c.Printf("%v\n", s.GetCardId())
			} else {
				c.Printf("%v - %v\n", s.GetCardId(), name)
			}
		}
	}
}

func setName(c *ishell.Context) {
	if ready := checkActiveCard(c); !ready {
		return
	}

	numCorrectArgs := 1
	if len(c.Args) != numCorrectArgs {
		c.Printf("setName requires %v args\n", numCorrectArgs)
		return
	}

	name := c.Args[0]
	err := activeCard.SetName(name)
	if err != nil {
		c.Printf("error setting name: %v", err)
	} else {
		updatePrompt()
		c.Printf("name set to %v", name)
	}
}

func activateCard(c *ishell.Context) {
	sessions := t.ListSessions()
	var sessionNames []string
	for _, session := range sessions {
		sessionNames = append(sessionNames, getDisplayName(session))
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
	if len(c.Args) != 1 {
		fmt.Println("wrong number of arguments given")
		return
	}
	CounterPartyConnInfo := c.Args[0]
	err := activeCard.ConnectToRemoteProvider(CounterPartyConnInfo)
	if err != nil {
		c.Err(err)
	}
}

func connectLocalSession(c *ishell.Context) {
	fmt.Println("Connecting to local counterparty provider")
	err := activeCard.ConnectToLocalProvider()
	if err != nil {
		c.Err(err)
	}
}

func pairWithCounterparty(c *ishell.Context) {
	c.Println("Pairing with card")
	if len(c.Args) != 1 {
		fmt.Println("wrong number of arguments given")
		return
	}
	CounterPartyID := strings.TrimSpace(c.Args[0])

	err := activeCard.ConnectToCounterparty(CounterPartyID)
	if err != nil {
		c.Err(err)
	}
}
