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
	"github.com/spf13/cobra"
)

// transactionAckCmd represents the transactionAck command
var transactionAckCmd = &cobra.Command{
	Use:   "transactionAck",
	Short: "Acknowledge a phonon transaction has been completed.",
	Long: `Acknowledge a phonon transaction has been completed so that
	the card can clean up intermediate data that may have been held during
	the transaction session`,
	Run: func(cmd *cobra.Command, args []string) {
		transactionAck()
	},
}

func init() {
	rootCmd.AddCommand(transactionAckCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transactionAckCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transactionAckCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

//Creates a phonon, sends it, and confirms the transaction for testing purposes
func transactionAck() {
	cs, err := card.OpenSecureConnection()
	if err != nil {
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	keyIndex, _, err := cs.CreatePhonon()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("created phonon with keyIndex: ", keyIndex)
	err = cs.SetDescriptor(&model.Phonon{KeyIndex: keyIndex, CurrencyType: model.Bitcoin, Denomination: 1})
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = cs.SendPhonons([]uint16{keyIndex}, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cs.TransactionAck([]uint16{keyIndex})
	if err != nil {
		fmt.Println(err)
		return
	}
}
