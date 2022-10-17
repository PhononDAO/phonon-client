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

// getFriendlyNameCmd represents the GetFriendlyName command
var getFriendlyNameCmd = &cobra.Command{
	Use:   "getFriendlyName",
	Short: "retrieve the card's previously set Friendly Name",
	Run: func(_ *cobra.Command, _ []string) {
		getFriendlyName()
	},
}

func init() {
	rootCmd.AddCommand(getFriendlyNameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// GetFriendlyNameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// GetFriendlyNameCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getFriendlyName() {
	conf := config.MustLoadConfig()
	cs, err := card.QuickSecureConnection(readerIndex, staticPairing, conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		fmt.Println("unable to verify pin: ", err)
		return
	}
	fmt.Println("setting 160 byte name")
	tooLongString := "AAAAAalsdkfja;lihnkjanhgpioauhrtoewknvkzngasoitruewaonkngvsadtuihewopirfjfkvndsaklfhaiulorhewpirjkfnjasdf;sdnf;lasdfjpoewirwnvnvgaotiweprouwr"
	err = cs.SetFriendlyName(tooLongString)
	if err == nil {
		fmt.Println("first set should have failed but did not")
		return
	}
	friendlyName, err := cs.GetFriendlyName()
	if err != nil {
		fmt.Println("error in first getFriendlyName: ", err)
	} else {
		fmt.Println("got name: ", friendlyName)
	}
	nameLength33 := util.RandomKey(33)

	fmt.Println("setting 33 byte name")
	err = cs.SetFriendlyName(string(nameLength33))
	if err != nil {
		fmt.Println("33 bytes SetFriendlyName failed: ", err)
	}

	friendlyName, err = cs.GetFriendlyName()
	if err != nil {
		fmt.Println("error getting name after setting 33 bytes: ", err)
	} else {
		fmt.Println("got name: ", friendlyName)
	}
	fmt.Println("setting 8 byte name")
	err = cs.SetFriendlyName("testName")
	if err != nil {
		fmt.Println("setFriendly failed: ", err)
	}
	friendlyName, err = cs.GetFriendlyName()
	if err != nil {
		fmt.Println("error in second getFriendlyName: ", err)
	} else {
		fmt.Println("got name: ", friendlyName)
	}
}
