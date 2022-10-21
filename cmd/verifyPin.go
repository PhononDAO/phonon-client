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
	"github.com/GridPlus/phonon-client/util"
	"github.com/spf13/cobra"
)

// verifyPinCmd represents the verifyPin command
var verifyPinCmd = &cobra.Command{
	Use:   "verifyPin",
	Short: "Test pin verification",
	Long:  `Tests pin verification. Accepts PIN via secure cprompt`,
	Run: func(_ *cobra.Command, _ []string) {
		verifyPin()
	},
}

func init() {
	rootCmd.AddCommand(verifyPinCmd)
}

func verifyPin() {
	conf := config.MustLoadConfig()
	cs, err := card.QuickSecureConnection(readerIndex, staticPairing, conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	pin, err := util.PinPrompt()
	if err != nil {
		fmt.Println("error receiving pin")
		return
	}
	err = cs.VerifyPIN(pin)
	if err != nil {
		fmt.Println("unable to verify pin: ", err)
		return
	}
}
