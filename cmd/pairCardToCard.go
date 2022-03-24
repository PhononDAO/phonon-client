/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/GridPlus/phonon-client/orchestrator"

	"github.com/spf13/cobra"
)

// pairCardToCardCmd represents the pairCardToCard command
var pairCardToCardCmd = &cobra.Command{
	Use:   "pairCardToCard",
	Short: "Establish a pairing between 2 phonon cards",
	Long: `Establish a local pairing between 2 phonon cards connected via
	2 different card readers attached to the the same phonon-client terminal`,
	Run: func(cmd *cobra.Command, args []string) {
		PairCardToCard()
	},
}

var (
	useMockReceiver     bool
	useMockSender       bool
	senderReaderIndex   int
	receiverReaderIndex int
)

func init() {
	rootCmd.AddCommand(pairCardToCardCmd)

	pairCardToCardCmd.Flags().BoolVarP(&useMockReceiver, "mock-receiver", "m", false, "Use the mock card implementation as the receiver")
	pairCardToCardCmd.Flags().BoolVarP(&useMockSender, "mock-sender", "n", false, "Use the mock card implementation as the sender")

	pairCardToCardCmd.Flags().IntVarP(&receiverReaderIndex, "receiver-reader-index", "r", 0, "pass the reader index to use for the receiver card")
	pairCardToCardCmd.Flags().IntVarP(&senderReaderIndex, "sender-reader-index", "s", 0, "pass the reader index to use for the sender card")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pairCardToCardCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pairCardToCardCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func PairCardToCard() {
	fmt.Println("opening session with sender Card")
	var sender *orchestrator.Session
	var err error
	terminal := orchestrator.NewPhononTerminal()
	var sessions []*orchestrator.Session
	if !useMockSender || !useMockReceiver {
		sessions, err = terminal.RefreshSessions()
		if err != nil {
			log.Fatal("Unable to find connected card/s" + err.Error())
		}

	}
	if !useMockSender && senderReaderIndex > len(sessions)-1 {
		log.Fatal("not enough connected cards for senderReaderIndex:" + fmt.Sprint(senderReaderIndex))
	}
	if !useMockReceiver && receiverReaderIndex > len(sessions)-1 {
		log.Fatal("not enough connected cards for receiverReaderIndex:" + fmt.Sprint(receiverReaderIndex))
	}

	if useMockSender {
		senderCardID, err := terminal.GenerateMock()
		if err != nil {
			log.Fatal("Unable to generate mock sender: " + err.Error())
		}
		sender = terminal.SessionFromID(senderCardID)
	} else {
		sender = sessions[senderReaderIndex]
	}
	fmt.Println("sender verify PIN")
	err = sender.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	var receiverSession *orchestrator.Session
	if useMockReceiver {
		receiverCardID, err := terminal.GenerateMock()
		if err != nil {
			log.Fatal("Unable to generate mock receiver: " + err.Error())
		}
		receiverSession = terminal.SessionFromID(receiverCardID)
	} else {
		fmt.Println("opening physical connection with receiver card")
		receiverSession = sessions[receiverReaderIndex]
	}

	fmt.Println("verifying receiver PIN")
	err = receiverSession.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("starting card to card pairing")
	err = sender.ConnectToLocalProvider()
	if err != nil {
		fmt.Println("Unable to initialize local counterparty provider")
		fmt.Println(err.Error())
		return
	}
	err = receiverSession.ConnectToLocalProvider()
	if err != nil {
		fmt.Println("Unable to initialize local counterparty provider")
		fmt.Println(err.Error())
		return
	}

	err = sender.ConnectToCounterparty(receiverSession.GetName())
	if err != nil {
		fmt.Println("error during pairing with counterparty")
		fmt.Println(err)
		return
	}
	fmt.Println("cards paired succesfully!")
}
