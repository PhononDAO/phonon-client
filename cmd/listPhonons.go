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
	"github.com/spf13/cobra"
)

// listPhononsCmd represents the listPhonons command
var listPhononsCmd = &cobra.Command{
	Use:   "listPhonons",
	Short: "Lists all phonons on card matching filter",
	Long: `Lists all phonons on card matching the given filter, and returning the available phonon descriptor
	`,
	Args: cobra.ExactValidArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		listPhonons()
	},
}

var currency uint16
var greaterThanValue float32
var lessThanValue float32

func init() {
	rootCmd.AddCommand(listPhononsCmd)

	listPhononsCmd.PersistentFlags().Uint16VarP(&currency, "currencyType", "c", 0, "0 matches all phonons, 1 for Bitcoin, 2 for Ethereum")
	listPhononsCmd.PersistentFlags().Float32VarP(&greaterThanValue, "gt", "g", 0, "phonon denomination must be greater than this float32 value")
	listPhononsCmd.PersistentFlags().Float32VarP(&lessThanValue, "lt", "l", 0, "phonon denomination must be less than this float32 value")
}

func listPhonons() {
	conf := config.MustLoadConfig()
	cs, err := card.QuickSecureConnection(readerIndex, staticPairing, conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = cs.VerifyPIN("111111"); err != nil {
		fmt.Println(err)
		return
	}

	phonons, err := cs.ListPhonons(model.CurrencyType(currency), 0, 0, false)
	if err != nil {
		fmt.Println("error listing phonons: ", err)
		return
	}
	if len(phonons) == 0 {
		fmt.Println("no phonons found")
		return
	}
	fmt.Println("retrieved phonons: ")
	for _, phonon := range phonons {
		fmt.Println(phonon)
	}
}
