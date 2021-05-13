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

// getPhononPubKeyCmd represents the getPhononPubKey command
var getPhononPubKeyCmd = &cobra.Command{
	Use:   "getPhononPubKey",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		getPhononPubKey()
	},
}

func init() {
	rootCmd.AddCommand(getPhononPubKeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getPhononPubKeyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getPhononPubKeyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getPhononPubKey() {
	cs, err := card.Connect()
	if err != nil {
		return
	}
	// cs, err := card.OpenSecureConnection()
	// if err != nil {
	// 	return
	// }
	// err = cs.VerifyPIN("111111")
	// if err != nil {
	// 	return
	// }
	keyIndex, _, err := cs.CreatePhonon()
	if err != nil {
		fmt.Println("error creating phonon: ", err)
		return
	}
	// keyIndex := uint16(1)
	err = cs.SetDescriptor(keyIndex, model.Bitcoin, 1)
	if err != nil {
		fmt.Println("error setting descriptor", err)
		return
	}
	pubKey, err := cs.GetPhononPubKey(keyIndex)
	if err != nil {
		fmt.Println("error getting phonon public key: ", err)
		return
	}
	fmt.Printf("got pubkey X: % X Y: % X\n", pubKey.X.Bytes(), pubKey.Y.Bytes())
}
