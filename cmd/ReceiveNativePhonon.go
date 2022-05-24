/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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
	"github.com/spf13/cobra"
)

// ReceiveNativePhononCmd represents the ReceiveNativePhonon command
var ReceiveNativePhononCmd = &cobra.Command{
	Use:   "ReceiveNativePhonon",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, _ = cmd, args // make my linter happy
		receiveNativePhonon()
	},
}

func init() {
	rootCmd.AddCommand(ReceiveNativePhononCmd)
	ReceiveNativePhononCmd.Flags().BoolVarP(&useMockReceiver, "mock-receiver", "m", false, "Use the mock card implementation as the receiver")
	ReceiveNativePhononCmd.Flags().BoolVarP(&useMockSender, "mock-sender", "n", false, "Use the mock card implementation as the sender")

	ReceiveNativePhononCmd.Flags().IntVarP(&receiverReaderIndex, "receiver-reader-index", "r", 0, "pass the reader index to use for the receiver card")
	ReceiveNativePhononCmd.Flags().IntVarP(&senderReaderIndex, "sender-reader-index", "s", 0, "pass the reader index to use for the sender card")
}

func receiveNativePhonon() {
	var senderCard model.PhononCard
	var err error
	if useMockSender {
		senderCard, err = card.NewMockCard(true, false)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		senderCard, err = card.QuickSecureConnection(senderReaderIndex, staticPairing)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	sender, _ := orchestrator.NewSession(senderCard)
	err = sender.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	var receiverCard model.PhononCard
	if useMockReceiver {
		receiverCard, err = card.NewMockCard(true, false)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		receiverCard, err = card.QuickSecureConnection(receiverReaderIndex, staticPairing)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	receiver, _ := orchestrator.NewSession(receiverCard)
	err = receiver.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	index, hash, err := senderCard.MineNativePhonon(uint8(1))
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("created native phonon hash: " + string(hash))
	orchestrator.NewPhononTerminal().AddSession(sender)
	orchestrator.NewPhononTerminal().AddSession(receiver)

	sender.ConnectToLocalProvider()
	receiver.ConnectToLocalProvider()

	err = sender.ConnectToCounterparty(receiver.GetName())
	if err != nil {
		panic(err.Error())
	}
	sender.SendPhonons([]uint16{index})

	phonons, err := receiver.ListPhonons(0, 0, 0)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%+v\n", phonons)
	for _, phonon := range phonons {
		k, err := receiver.GetPhononPubKey(phonon.KeyIndex, phonon.CurveType)
		if err != nil {
			panic(err.Error)
		}
		fmt.Printf("recieved public Key:% X \n", k.Bytes())
	}

}
