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
	"github.com/GridPlus/phonon-client/util"
	"github.com/spf13/cobra"
)

// changePinCmd represents the changePin command
var changePinCmd = &cobra.Command{
	Use:   "changePin",
	Short: "Change card's 6 digit PIN",
	Long: `Change the card's existing 6 digit PIN to a new one.
command will prompt you securely for the existing PIN as well as the new one.`,
	Run: func(_ *cobra.Command, _ []string) {
		changePin()
	},
}

func init() {
	rootCmd.AddCommand(changePinCmd)
}

func changePin() {
	cs, err := card.QuickSecureConnection(readerIndex, staticPairing)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("enter current pin for verification")
	verificationPin, err := util.PinPrompt()
	if err != nil {
		fmt.Println("error receiving pin")
		return
	}
	err = cs.VerifyPIN(verificationPin)
	if err != nil {
		fmt.Println("unable to verify pin: ", err)
		return
	}
	fmt.Println("enter new pin")
	pin, err := util.PinPrompt()
	if err != nil {
		fmt.Println("error receiving pin")
		return
	}
	err = cs.ChangePIN(pin)
	if err != nil {
		fmt.Println("unable to change pin: ", err)
		return
	}

}
