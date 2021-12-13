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
	"crypto/ecdsa"
	"fmt"

	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/orchestrator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// setReceiveListCmd represents the setReceiveList command
var setReceiveListCmd = &cobra.Command{
	Use:   "setReceiveList",
	Short: "low level test of SET_RECV_LIST command",
	Long:  `low level test of SET_RECV_LIST command`,
	Run: func(cmd *cobra.Command, args []string) {
		setReceiveList()
	},
}

func init() {
	rootCmd.AddCommand(setReceiveListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setReceiveListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setReceiveListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setReceiveList() {
	cs, err := orchestrator.QuickSecureConnection(readerIndex)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Create a phonon, get it's pubKey, and then set it in the RECV_LIST for testing
	_, pubKey, err := cs.CreatePhonon(model.Secp256k1)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cs.SetReceiveList([]*ecdsa.PublicKey{pubKey})
	if err != nil {
		log.Error("error testing SetReceiveList: ", err)
		return
	}
}
