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
	"github.com/GridPlus/phonon-client/model"
	"github.com/spf13/cobra"
)

// createPhononCmd represents the createPhonon command
var createPhononCmd = &cobra.Command{
	Use:   "createPhonon",
	Short: "Create a new phonon",
	Long: `Creates a new phonon returning the public key and current keyIndex,
an identifier which is valid for the duration of a card session. KeyIndices may change
when the SELECT command is run against the card again.

Phonons created by this command have no identifying descriptor information.
`,
	Run: func(cmd *cobra.Command, args []string) {
		var n int
		if len(args) < 1 {
			n = 1
		} else {
			var err error
			if n, err = strconv.Atoi(args[0]); err != nil {
				fmt.Println("argument must be an integer")
				return
			}
		}
		createPhonon(n)
	},
}

func init() {
	rootCmd.AddCommand(createPhononCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createPhononCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createPhononCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createPhonon(n int) {
	cs, err := card.OpenSecureConnection()
	if err != nil {
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		return
	}
	for i := 0; i < n; i++ {
		keyIndex, pubKey, err := cs.CreatePhonon(model.Secp256k1)
		if err != nil {
			fmt.Println("error creating phonon")
			fmt.Println(err)
			return
		}
		fmt.Printf("created phonon with keyIndex %v and pubKey % X\n", keyIndex, append(pubKey.X.Bytes(), pubKey.Y.Bytes()...))
	}
}
