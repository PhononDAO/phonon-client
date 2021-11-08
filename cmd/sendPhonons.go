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
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/spf13/cobra"
)

// sendPhononsCmd represents the sendPhonons command
var sendPhononsCmd = &cobra.Command{
	Use:   "sendPhonons",
	Short: "Send a packet of phonons",
	Long:  `Send a packet of phonons with a list of keyIndices,`,
	Run: func(cmd *cobra.Command, args []string) {
		sendPhonons()
	},
}

func init() {
	rootCmd.AddCommand(sendPhononsCmd)

	sendPhononsCmd.Flags().BoolVarP(&useMockReceiver, "mock-receiver", "m", false, "Use the mock card implementation as the receiver")
	sendPhononsCmd.Flags().BoolVarP(&useMockSender, "mock-sender", "n", false, "Use the mock card implementation as the sender")

	sendPhononsCmd.Flags().IntVarP(&receiverReaderIndex, "receiver-reader-index", "r", 0, "pass the reader index to use for the receiver card")
	sendPhononsCmd.Flags().IntVarP(&senderReaderIndex, "sender-reader-index", "s", 0, "pass the reader index to use for the sender card")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendPhononsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendPhononsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sendPhonons() {
	fmt.Println("opening session with sender Card")
	var senderCard model.PhononCard
	var err error
	if useMockSender {
		senderCard, err = card.NewMockCard()
		if err != nil {
			fmt.Println(err)
			return
		}
		senderCard.InstallCertificate(cert.SignWithDemoKey)
		senderCard.Init("111111")
	} else {
		senderCard, _, err = card.OpenBestConnectionWithReaderIndex(senderReaderIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	sender, _ := card.NewSession(senderCard)
	err = sender.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}

	//Create a single phonon to transfer

	keyIndex, _, err := sender.CreatePhonon()
	if err != nil {
		fmt.Println(err)
		return
	}
	p := &model.Phonon{
		KeyIndex:     keyIndex,
		CurrencyType: model.Ethereum,
		Denomination: 1,
	}
	err = sender.SetDescriptor(p)
	if err != nil {
		fmt.Println(err)
		return
	}

	phonons, err := sender.ListPhonons(model.Ethereum, 0, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, p := range phonons {
		fmt.Println("listing sender phonons: ")
		fmt.Printf("%+v\n", p)
	}
	var receiverCard model.PhononCard
	if useMockReceiver {
		receiverCard, err = card.NewMockCard()
		if err != nil {
			fmt.Println(err)
			return
		}
		receiverCard.InstallCertificate(cert.SignWithDemoKey)
		receiverCard.Init("111111")
	} else {
		receiverCard, _, err = card.OpenBestConnectionWithReaderIndex(receiverReaderIndex)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("opening receiver session")
	receiverSession, _ := card.NewSession(receiverCard)
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

	err = sender.SendPhonons([]uint16{keyIndex})
	if err != nil {
		fmt.Println("error sending phonons")
		fmt.Println(err)
		return
	}

	fmt.Println("sent phonons without error")
	phonons, err = receiverSession.ListPhonons(model.Ethereum, 0, 0)
	if err != nil {
		fmt.Println("unable to list receiver phonons: ", err)
		return
	}
	for _, p := range phonons {
		fmt.Println("receiver has phonons: ")
		fmt.Printf("%+v", p)
	}
}
