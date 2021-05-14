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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		setDescriptor()
	},
}

func init() {
	rootCmd.AddCommand(setDescriptorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setDescriptorCmd.PersistentFlags().String("foo", "", "A help for foo")

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

	keyIndex, _, err := cs.CreatePhonon()
	if err != nil {
		fmt.Println("error creating phonon: ", err)
		return
	}

	fmt.Println("sending set descriptor")
	//Create a mock BTC descriptor
	err = cs.SetDescriptor(keyIndex, model.Ethereum, 100)
	if err != nil {
		fmt.Println("unable to set descriptor")
		return
	}

	//Create a mock ETH descriptor
}
