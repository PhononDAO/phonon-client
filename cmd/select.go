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

// selectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Test SELECT command",
	Long:  `Test SELECT command`,
	Run: func(cmd *cobra.Command, args []string) {
		cs, err := card.Connect(readerIndex)
		if err != nil {
			return
		}
		_, _, _, err = cs.Select()
		if err != nil && err != card.ErrCardUninitialized {
			fmt.Println("could not select applet", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
}
