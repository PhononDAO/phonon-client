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

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [pin]",
	Short: "Initialize phonon card with PIN.",
	Long:  `Initialize phonon card with PIN. Defaults to 111111 if no argument is given`,
	Run: func(cmd *cobra.Command, args []string) {
		var pin string
		if len(args) > 0 {
			pin = args[0]
		} else {
			pin = "111111"
		}
		initializeCard(pin)
	},
	Args: cobra.MaximumNArgs(1),
}

var readerIndex int

func init() {
	rootCmd.AddCommand(initCmd)

	pairCardToCardCmd.Flags().IntVarP(&readerIndex, "reader-index", "i", 0, "pass the reader index for the card")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

//TODO: reimplement with new Session
func initializeCard(pin string) {
	fmt.Println("running initializeCard!!!")

	cs, _ := card.ConnectWithReaderIndex(readerIndex)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	fmt.Println("cmdline is going to select")
	// s := card.NewSession(cs, false)
	_, _, _, _ = cs.Select()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	err := cs.Init(pin)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("successfully set PIN")
}
