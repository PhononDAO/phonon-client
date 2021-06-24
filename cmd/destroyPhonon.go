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

// destroyPhononCmd represents the destroyPhonon command
var destroyPhononCmd = &cobra.Command{
	Use:   "destroyPhonon [keyIndex]",
	Short: "Destroy a phonon by keyIndex",
	Long: `Destroy a phonon by it's keyIndex, returning the private key.

This allows one to utilize the phonon's private key outside of the phonon system,
but the phonon will no longer be retrievable via the card.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyIndex, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			fmt.Println("couldn't parse keyIndex value as uint16: ", err)
			return
		}
		destroyPhonon(uint16(keyIndex))
	},
}

func init() {
	rootCmd.AddCommand(destroyPhononCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyPhononCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyPhononCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func destroyPhonon(keyIndex uint16) {
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

	privKey, err := cs.DestroyPhonon(keyIndex)
	if err != nil {
		return
	}
	fmt.Println("destroyed phonon and exported privKey: ")
	fmt.Printf("D: % X", privKey.D)
}
