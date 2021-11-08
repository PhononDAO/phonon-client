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
	"crypto/ecdsa"
	"fmt"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/model"
	"github.com/spf13/cobra"
)

// setListAndReceive represents the setAndReceiveList command
var setListAndReceiveCmd = &cobra.Command{
	Use:   "setListAndReceive",
	Short: "Tests SET_RECV_LIST and RECV_PHONONS functionality",
	Long: `Tests SET_RECV_LIST AND RECV_PHONONS functionality with a single card by doing the following.

	1. Creates a phonon.
	2. Sets Descriptor on that phonon
	3. Sends SET_RECV_LIST referring to that same phonon.
	4. Asks the card to send the phonon.
	5. Asks the card to receive back the same phonon, now that it is validated by SET_RECV_LIST

	`,
	Run: func(cmd *cobra.Command, args []string) {
		setListAndReceive()
	},
}

func init() {
	rootCmd.AddCommand(setListAndReceiveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setListAndReceive.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setListAndReceive.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setListAndReceive() {
	cs, err := card.OpenSecureConnection()
	if err != nil {
		return
	}

	err = cs.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	keyIndex, phononPubKey, err := cs.CreatePhonon()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cs.SetDescriptor(&model.Phonon{KeyIndex: uint16(keyIndex), CurrencyType: model.Ethereum, Denomination: 5})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("created phonon with keyIndex % X\n", keyIndex)
	phononTransfer, err := cs.SendPhonons([]uint16{uint16(keyIndex)}, false)
	if err != nil {
		fmt.Println("error sending phonon: ", err)
		return
	}

	//TODO: actually test this
	// badTransfer, err := cs.SendPhonons([]uint16{uint16(1)}, false)
	// if err != nil {
	// 	return
	// }
	err = cs.SetReceiveList([]*ecdsa.PublicKey{phononPubKey})
	if err != nil {
		return
	}
	//bad request
	err = cs.ReceivePhonons(phononTransfer)
	//good request
	// err = cs.ReceivePhonons(transferPackets[0])
	if err != nil {
		return
	}

	err = cs.TransactionAck([]uint16{keyIndex})
	if err != nil {
		return
	}
}
