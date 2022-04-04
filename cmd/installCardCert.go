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
	"bufio"
	"fmt"
	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/model"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// installCardCert represents the installCardCert command
var installCardCert = &cobra.Command{
	Use:   "installCardCert",
	Short: "This is to sign and install an identity certificate to the card",
	Run: func(cmd *cobra.Command, args []string) {
		InstallCardCert()
	},
}

var (
	useDemoKey    bool
	provisionMode bool
	yubikeySlot   string
	yubikeyPass   string
)

func init() {
	rootCmd.AddCommand(installCardCert)

	installCardCert.Flags().BoolVarP(&useDemoKey, "demo", "d", false, "Use the demo key to sign -- insecure for demo purposes only")
	installCardCert.Flags().BoolVarP(&provisionMode, "provision", "p", false, "suppress all output except for the cardID for automated provisioning")
	installCardCert.Flags().StringVarP(&yubikeySlot, "slot", "s", "", "Slot in which the signing yubikey is insterted") //this is taken in as a string to allow for a nil value instead of 0 value
	installCardCert.Flags().StringVarP(&yubikeyPass, "pass", "", "", "Yubikey Password")
}

func InstallCardCert() {
	var signKeyFunc func([]byte) ([]byte, error)
	// Determine Signing Key
	if useDemoKey {
		if !provisionMode {
			fmt.Println("Using Demo Key!")
		}
		signKeyFunc = cert.SignWithDemoKey
	} else {
		//gather information for ubikey signing
		var yubikeySlotInt int
		yubikeySlotInt, err := strconv.Atoi(yubikeySlot)
		if err != nil {
			fmt.Println("Please enter the yubikey slot: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			input = strings.TrimSuffix(input, "\n")
			yubikeySlotInt, err = strconv.Atoi(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if yubikeyPass == "" {
			fmt.Println("Please enter the yubikey password:")
			passBytes, err := term.ReadPassword(0)
			if err != nil {
				log.Fatalf("Unable to retrieve password from console: %s", err.Error())
			}
			yubikeyPass = string(passBytes)
		}

		signKeyFunc = cert.SignWithYubikeyFunc(yubikeySlotInt, yubikeyPass)
	}

	var cs model.PhononCard
	baseCS, err := card.Connect(readerIndex)
	if err != nil {
		log.Fatalf("Unable to connect to card: %s", err.Error())
	}
	if staticPairing {
		cs = card.NewStaticPhononCommandSet(baseCS)
	} else {
		cs = baseCS
	}
	_, cardPubKey, _, err := cs.Select()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cs.InstallCertificate(signKeyFunc)
	if err != nil {
		log.Fatalf("Unable to Install Certificate: %s", err.Error())
	}
	fmt.Printf("%x\n", ethcrypto.FromECDSAPub(cardPubKey))
}
