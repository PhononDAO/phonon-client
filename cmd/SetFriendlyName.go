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
	"github.com/spf13/cobra"
)

var name string

// setFriendlyNameCmd represents the SetFriendlyName command
var setFriendlyNameCmd = &cobra.Command{
	Use:   "setFriendlyName",
	Short: "Set friendly name for phonon card",
	Run: func(_ *cobra.Command, _ []string) {
		setFriendlyName()
	},
}

func init() {
	setFriendlyNameCmd.PersistentFlags().StringVar(&name, "name", "", "Friendly name to call your card")
	rootCmd.AddCommand(setFriendlyNameCmd)
}

func setFriendlyName() {
	if name == "" {
		fmt.Println("Please input a name to set for the card")
	}
	conf := config.MustLoadConfig()
	cs, err := card.QuickSecureConnection(readerIndex, staticPairing, conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		fmt.Println("can't verify pin")
		fmt.Println(err.Error())
		return
	}
	cs.SetFriendlyName(name)
}
