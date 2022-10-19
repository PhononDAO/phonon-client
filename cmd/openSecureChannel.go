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

// openSecureChannelCmd represents the openSecureChannel command
var openSecureChannelCmd = &cobra.Command{
	Use:   "openSecureChannel",
	Short: "Tests opening a secure channel",
	Long:  `Tests opening a secure channel between terminal and card`,
	Run: func(_ *cobra.Command, _ []string) {
		openSecureChannel()
	},
}

func init() {
	rootCmd.AddCommand(openSecureChannelCmd)
}

func openSecureChannel() {
	conf := config.MustLoadConfig()
	cs, err := card.Connect(readerIndex, conf)
	if err != nil {
		fmt.Println("could not connect to card: ", err)
	}
	_, _, _, err = cs.Select()
	if err != nil {
		fmt.Println("could not select phonon applet: ", err)
		return
	}
	_, err = cs.Pair()
	if err != nil {
		fmt.Println("could not pair: ", err)
		return
	}
	err = cs.OpenSecureChannel()
	if err != nil {
		fmt.Println("could not open secure channel: ", err)
		return
	}
	fmt.Println("secure channel opened without error")
}
