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

// setDescriptorCmd represents the setDescriptor command
var setDescriptorCmd = &cobra.Command{
	Use:   "setDescriptor",
	Short: "Set a description of a phonon",
	Long: `Set a phonon description in order to store metadata that identifies which blockchain assets
the phonon corresponds to, so that they can later be retrieved for use in transactions.`,
	Run: func(cmd *cobra.Command, args []string) {
		setDescriptor()
	},
}

var phononCount int

func init() {
	rootCmd.AddCommand(setDescriptorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setDescriptorCmd.PersistentFlags().String("foo", "", "A help for foo")
	setDescriptorCmd.PersistentFlags().IntVarP(&phononCount, "count", "c", 1, "number of phonons to create with descriptor set")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setDescriptorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setDescriptor() {
	cs, err := card.OpenSecureConnection()
	if err != nil {
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		return
	}
	for i := 0; i < phononCount; i++ {
		keyIndex, _, err := cs.CreatePhonon()
		if err != nil {
			fmt.Println("error creating phonon: ", err)
			return
		}

		fmt.Println("sending set descriptor for keyIndex ", keyIndex)
		err = cs.SetDescriptor(&model.Phonon{KeyIndex: keyIndex, CurrencyType: model.Ethereum, Denomination: 100})
		if err != nil {
			fmt.Println("unable to set descriptor")
			return
		}
	}
	fmt.Printf("successfully created %v phonons with descriptor(s)\n", phononCount)

}
