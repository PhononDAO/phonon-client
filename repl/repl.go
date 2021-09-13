package repl

import (
	"github.com/GridPlus/phonon-client/terminal"
	ishell "github.com/abiosoft/ishell/v2"
)

var t terminal.PhononTerminal

func Start() {
	shell := ishell.New()
	t = terminal.PhononTerminal{}

	shell.Println("Welcome to the phonon command interface")

	shell.AddCmd(&ishell.Cmd{
		Name:    "list",
		Aliases: []string{},
		Func:    listCards,
		Help:    "List available phonon cards for usage",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "unlock",
		Func: unlock,
		Help: "Unlock card with pin",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "listPhonons",
		Func: listPhonons,
		Help: "List phonons on card at index",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "create",
		Func: createPhonon,
		Help: "Create a phonon on card at index",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Func: setDescriptor,
		Help: "Set the type and value of phonon on card at index with value at index",
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "balance",
		Func: getBalance,
		Help: "Retrieve the type and balance of a phonon on card at index with value at index",
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
	shell.Run()
}

func listCards(c *ishell.Context) {
	t.ListSessions()
}

func unlock(c *ishell.Context) {
	var sessionIndex int
	var pin string
	t.UnlockCard(sessionIndex, pin)
}

func listPhonons(c *ishell.Context) {
	var sessionIndex int
	t.ListPhonons(sessionIndex)
}

func createPhonon(c *ishell.Context) {
	var sessionIndex int
	t.CreatePhonon(sessionIndex)
}

func setDescriptor(c *ishell.Context) {
	var sessionIndex int
	var phononIndex int
	t.SetDescriptor(sessionIndex, phononIndex, struct{}{})
}

func getBalance(c *ishell.Context) {
	var cardIndex, phononIndex int
	t.GetBalance(cardIndex, phononIndex)
}

func connectRemoteSession(c *ishell.Context) {
	var sessionIndex int
	t.ConnectRemoteSession(sessionIndex, struct{}{})
}

func setReceiveMode(c *ishell.Context) {
	var sessionIndex int
	t.SetReceiveMode(sessionIndex)
}

func redeemPhonon(c *ishell.Context) {
	var sessionIndex, phononIndex int
	t.RedeemPhonon(sessionIndex, phononIndex)
}
