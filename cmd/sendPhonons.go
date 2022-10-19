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
	"github.com/GridPlus/phonon-client/config"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/spf13/cobra"
)

// sendPhononsCmd represents the sendPhonons command
var sendPhononsCmd = &cobra.Command{
	Use:   "sendPhonons",
	Short: "Send a packet of phonons",
	Long:  `Send a packet of phonons with a list of keyIndices,`,
	Run: func(_ *cobra.Command, _ []string) {
		sendPhonons()
	},
}

func init() {
	rootCmd.AddCommand(sendPhononsCmd)

	sendPhononsCmd.Flags().BoolVarP(&useMockReceiver, "mock-receiver", "m", false, "Use the mock card implementation as the receiver")
	sendPhononsCmd.Flags().BoolVarP(&useMockSender, "mock-sender", "n", false, "Use the mock card implementation as the sender")

	sendPhononsCmd.Flags().IntVarP(&receiverReaderIndex, "receiver-reader-index", "r", 0, "pass the reader index to use for the receiver card")
	sendPhononsCmd.Flags().IntVarP(&senderReaderIndex, "sender-reader-index", "s", 0, "pass the reader index to use for the sender card")
}

func sendPhonons() {
	conf := config.MustLoadConfig()
	term := orchestrator.NewPhononTerminal(conf)

	fmt.Println("opening session with sender Card")
	var senderCard model.PhononCard
	var err error
	if useMockSender {
		senderCard, err = card.NewMockCard(true, false)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		senderCard, err = card.QuickSecureConnection(senderReaderIndex, staticPairing, conf)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	sender, _ := orchestrator.NewSession(senderCard)
	term.AddSession(sender)

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
		Denomination: model.Denomination{Base: 1, Exponent: 0},
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
		receiverCard, err = card.NewMockCard(true, false)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		receiverCard, err = card.QuickSecureConnection(receiverReaderIndex, staticPairing, conf)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("opening receiver session")
	receiver, _ := orchestrator.NewSession(receiverCard)
	term.AddSession(receiver)
	err = receiver.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = sender.ConnectToLocalProvider()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = receiver.ConnectToLocalProvider()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("starting card to card pairing")
	err = sender.ConnectToCounterparty(receiver.GetCardId())
	if err != nil {
		fmt.Println("error pairing sender to receiver")
		fmt.Println(err)
		return
	}
	err = receiver.ConnectToCounterparty(sender.GetCardId())
	if err != nil {
		fmt.Println("error pairing receiver to sender")
		return
	}
	fmt.Println("cards paired succesfully!")

	err = sender.SendPhonons([]model.PhononKeyIndex{keyIndex})
	if err != nil {
		fmt.Println("error sending phonons")
		fmt.Println(err)
		return
	}

	fmt.Println("sent phonons without error")
	phonons, err = receiver.ListPhonons(model.Ethereum, 0, 0)
	if err != nil {
		fmt.Println("unable to list receiver phonons: ", err)
		return
	}
	fmt.Println("receiver has phonons: ")
	for _, p := range phonons {
		fmt.Printf("%+v", p)
	}
}
