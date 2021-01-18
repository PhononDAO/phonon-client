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
	"github.com/spf13/cobra"
)

// openSecureChannelCmd represents the openSecureChannel command
var openSecureChannelCmd = &cobra.Command{
	Use:   "openSecureChannel",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		openSecureChannel()
	},
}

func init() {
	rootCmd.AddCommand(openSecureChannelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// openSecureChannelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// openSecureChannelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func openSecureChannel() {
	cs, err := card.Connect()
	if err != nil {
		fmt.Println("could not connect to card: ", err)
	}
	_, _, err = cs.Select()
	if err != nil {
		fmt.Println("could not select phonon applet: ", err)
	}
	err = cs.Pair()
	if err != nil {
		fmt.Println("could not pair: ", err)
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		fmt.Println("could not open secure channel: ", err)
	}
}