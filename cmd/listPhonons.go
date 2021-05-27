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

// listPhononsCmd represents the listPhonons command
var listPhononsCmd = &cobra.Command{
	Use:   "listPhonons",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		listPhonons()
	},
}

func init() {
	rootCmd.AddCommand(listPhononsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listPhononsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listPhononsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func listPhonons() {
	cs, err := card.OpenSecureConnection()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = cs.VerifyPIN("111111"); err != nil {
		fmt.Println(err)
		return
	}

	phonons, err := cs.ListPhonons(model.Unspecified, 0, 0)
	if err != nil {
		return
	}
	for _, phonon := range phonons {
		fmt.Printf("retrieved phonon: %+v\n", phonon)
	}
}
