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

// getPhononPubKeyCmd represents the getPhononPubKey command
var getPhononPubKeyCmd = &cobra.Command{
	Use:   "getPhononPubKey",
	Short: "Retrieves a phonon's public key by keyIndex",
	Long: `Retrieves a phonon's public key by keyIndex. This may be necessary to follow up
a list command, as LIST_PHONONS does not return public keys in order to conserve space.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyIndex, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		getPhononPubKey(uint16(keyIndex))
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

func getPhononPubKey(keyIndex uint16) {
	cs, err := card.OpenSecureConnection()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		fmt.Println(err)
		return
	}
	pubKey, err := cs.GetPhononPubKey(keyIndex)
	if err != nil {
		fmt.Println("error getting phonon public key: ", err)
		return
	}
	fmt.Printf("got pubkey X: % X Y: % X\n", pubKey.X.Bytes(), pubKey.Y.Bytes())
}
