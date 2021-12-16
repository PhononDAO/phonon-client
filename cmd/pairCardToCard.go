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

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/GridPlus/phonon-client/session"

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
	staticPairing       bool
)

func init() {
	rootCmd.AddCommand(pairCardToCardCmd)

	pairCardToCardCmd.Flags().BoolVarP(&useMockReceiver, "mock-receiver", "m", false, "Use the mock card implementation as the receiver")
	pairCardToCardCmd.Flags().BoolVarP(&useMockSender, "mock-sender", "n", false, "Use the mock card implementation as the sender")

	pairCardToCardCmd.Flags().IntVarP(&receiverReaderIndex, "receiver-reader-index", "r", 0, "pass the reader index to use for the receiver card")
	pairCardToCardCmd.Flags().IntVarP(&senderReaderIndex, "sender-reader-index", "s", 0, "pass the reader index to use for the sender card")

	pairCardToCardCmd.Flags().BoolVarP(&staticPairing, "static", "t", false, "Use statically generated insecure keys and salts to generate deterministic pairing payloads")

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
	var senderCard model.PhononCard
	var sender *session.Session
	var err error
	if useMockSender {
		senderCard, err := card.NewMockCard(true, staticPairing)
		if err != nil {
			fmt.Println(err)
			return
		}
		sender, err = session.NewSession(senderCard)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		senderCard, err = orchestrator.Connect(senderReaderIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
		sender, err = session.NewSession(senderCard)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println("sender verify PIN")
	err = sender.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	var receiverCard model.PhononCard
	var receiverSession *session.Session
	if useMockReceiver {
		receiverCard, err = card.NewMockCard(true, staticPairing)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("opening receiver session")
		receiverSession, err = session.NewSession(receiverCard)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("opening physical connection with receiver card")
		receiverCard, err = orchestrator.Connect(receiverReaderIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
		receiverSession, err = session.NewSession(receiverCard)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("verifying receiver PIN")
	err = receiverSession.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}

	receiver := orchestrator.NewLocalCounterParty(receiverSession)

	fmt.Println("starting card to card pairing")
	err = sender.PairWithRemoteCard(receiver)
	if err != nil {
		fmt.Println("error during pairing with counterparty")
		fmt.Println(err)
		return
	}
	fmt.Println("cards paired succesfully!")
}
