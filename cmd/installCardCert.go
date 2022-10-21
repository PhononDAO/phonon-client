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
	"log"
	"os"
	"strings"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/config"
	"github.com/GridPlus/phonon-client/model"
	"github.com/GridPlus/phonon-client/util"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// installCardCert represents the installCardCert command
var installCardCert = &cobra.Command{
	Use:   "installCardCert",
	Short: "This is to sign and install an identity certificate to the card",
	Run: func(_ *cobra.Command, _ []string) {
		InstallCardCert()
	},
}

var (
	useDemoKey    bool
	provisionMode bool
	yubikeySlot   int
	yubikeyPass   string
)

func init() {
	rootCmd.AddCommand(installCardCert)

	installCardCert.Flags().BoolVarP(&useDemoKey, "demo", "d", false, "Use the demo key to sign -- insecure for demo purposes only")
	installCardCert.Flags().BoolVarP(&provisionMode, "provision", "p", false, "suppress all output except for the cardID for automated provisioning")
	installCardCert.Flags().IntVarP(&yubikeySlot, "slot", "s", 0, "Slot in which the signing yubikey is insterted")
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
		//gather information for yubikey signing

		if yubikeySlot == 0 {
			fmt.Println("Please enter the yubikey slot: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			input = strings.TrimSuffix(input, "\n")
		}
		if yubikeyPass == "" {
			fmt.Println("Please enter the yubikey password:")
			passBytes, err := term.ReadPassword(0)
			if err != nil {
				log.Fatalf("Unable to retrieve password from console: %s", err.Error())
			}
			yubikeyPass = string(passBytes)
		}

		signKeyFunc = cert.SignWithYubikeyFunc(yubikeySlot, yubikeyPass)
	}

	var cs model.PhononCard
	conf := config.MustLoadConfig()
	baseCS, err := card.Connect(readerIndex, conf)
	if err != nil {
		log.Fatalf("Unable to connect to card: %s", err.Error())
	}
	if staticPairing {
		cs = card.NewStaticPhononCommandSet(baseCS)
	} else {
		cs = baseCS
	}
	_, _, _, err = cs.Select()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cs.InstallCertificate(signKeyFunc)
	if err != nil {
		log.Fatalf("Unable to Install Certificate: %s", err.Error())
	}
	key, _, err := cs.IdentifyCard([]byte{0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03, 0x00, 0x01, 0x02, 0x03})
	if err != nil {
		log.Fatal("Unable to connect to card: " + err.Error())
	}
	fmt.Printf("%s\n", util.CardIDFromPubKey(key))
}
