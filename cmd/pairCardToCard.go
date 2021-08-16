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
	"strconv"

	"github.com/GridPlus/phonon-client/card"
	"github.com/spf13/cobra"
)

// pairCardToCardCmd represents the pairCardToCard command
var pairCardToCardCmd = &cobra.Command{
	Use:   "pairCardToCard",
	Short: "Establish a pairing between 2 phonon cards",
	Long: `Establish a local pairing between 2 phonon cards connected via
	2 different card readers attached to the the same phonon-client terminal`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		senderReaderIndex, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
		}
		receiverReaderIndex, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
		}
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

	pairCardToCardCmd.Flags().BoolVarP(&useMockReceiver, "mock-receiver", "r", false, "Use the mock card implementation as the receiver")
	pairCardToCardCmd.Flags().BoolVarP(&useMockSender, "mock-sender", "s", false, "Use the mock card implementation as the sender")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pairCardToCardCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pairCardToCardCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

//TODO: Make flags to intelligently specify reader indices
func PairCardToCard() {
	fmt.Println("opening session with sender Card")
	var senderCard card.PhononCard
	var err error
	if useMockSender {
		senderCard, err = card.NewMockCard()
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		senderCard, _, err = card.OpenBestConnectionWithReaderIndex(receiverReaderIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	sender := card.NewSession(senderCard, true)

	var receiverCard card.PhononCard
	if useMockReceiver {
		receiverCard, err = card.NewMockCard()
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		receiverCard, _, err = card.OpenBestConnectionWithReaderIndex(receiverReaderIndex)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("opening receiver session")
	receiverSession := card.NewSession(receiverCard, true)

	receiver := card.NewLocalCounterParty(receiverSession)

	fmt.Println("starting card to card pairing")
	err = sender.PairWithRemoteCard(receiver)
	if err != nil {
		fmt.Println("error during pairing with counterparty")
		fmt.Println(err)
		return
	}
}
