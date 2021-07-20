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
	"crypto/rand"
	"fmt"
	"io"
	"log"

	"github.com/GridPlus/phonon-client/card"
	"github.com/spf13/cobra"
)

// installCardCertCmd represents the installCardCert command
var installCardCertCmd = &cobra.Command{
	Use:   "Install Certificate to Card",
	Short: "This is to sign and install a certificate to the card",
	Long:  `This is a longer version of saying it installs a ceritificate to the card signed by the authority`,
	Run: func(cmd *cobra.Command, args []string) {
		DoTheThing()
	},
}

var (
	useDemoKey      bool
	ubikeySlot      int
	ubikeyPass      string
	usePhononApplet bool
)

func init() {
	rootCmd.AddCommand(installCardCertCmd)

	installCardCertCmd.Flags().BoolVarP(&useDemoKey, "demo", "d", false, "Use the demo key to sign -- insecure for demo purposes only")
	installCardCertCmd.Flags().IntVarP(&ubikeySlot, "slot", "s", 0, "Slot in which the signing ubikey is insterted")
	installCardCertCmd.Flags().StringVarP(&ubikeyPass, "pass", "", "", "Ubikey Password")
	installCardCertCmd.Flags().BoolVarP(&usePhononApplet, "Phonon", "p", false, "install certificate on a phonon applet, instead of the safecard applet")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCardCertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCardCertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func DoTheThing() {
	var signKeyFunc func([]byte)([]byte, error)
	// Determine Signing Key
	if useDemoKey {
		fmt.Println("Using Demo Key!")
		signKeyFunc = SignWithDemoKey
	} else {
		//todo: implement logic with yubikey
	}
	// get card readers
	index := 0 //todo: get card reader index

	// Select Card
	cs, err := card.ConnectWithReaderIndex(index) //todo: make this a secure session
	if err != nil {
		log.Fatalf("Unable to connect to card %d: %s",index, err.Error())
	}

	nonce := make([]byte, 32)
	n, err := io.ReadFull(rand.Reader,nonce)
	if err != nil{
		log.Fatalf("Unable to retrieve random challenge to card")
	}
	if n != 32{
		log.Fatalf("Unable to read 32 bytes for challenge to card")
	}
	// Send Challenge to card
	cardPubKey, _, err := cs.IdentifyCard(nonce)
	if err != nil{
		log.Fatalf("Unable to Verify card %s", err.Error())
	}
	
	// make Card Certificate
	perms := []byte{0x30, 0x00, 0x02, 0x02, 0x00, 0x00, 0x80, 0x41} //todo: ask what this is
	cardCertificate := append(perms, cardPubKey...)

	// sign The Certificate
	preImage := cardCertificate[2:]
	
	sig, err := signKeyFunc(preImage)
	if err != nil{
		log.Fatalf("Unable to sign Cert: %s", err.Error())
	}
	// Append CA Signature to certificate
	signedCert := append(cardCertificate, sig...)
	// Install Certificate into Safecard applet
	err = cs.InstallCertificate(signedCert)
	if err != nil{
		log.Fatalf("Unable to install Certificate to card: %s", err.Error())
	}
	// Disconnect from card

	// release card reader
}

func SignWithUbikey() {
	//todo: this
}

func SignWithDemoKey(cert []byte) ([]byte, error) {
/*	demoKey := []byte{
		0x03, 0x8D, 0x01, 0x08, 0x90, 0x00, 0x00, 0x00,
		0x10, 0xAA, 0x82, 0x07, 0x09, 0x80, 0x00, 0x00,
		0x01, 0xBB, 0x03, 0x06, 0x90, 0x08, 0x35, 0xF9,
		0x10, 0xCC, 0x04, 0x85, 0x09, 0x00, 0x00, 0x91,
	}
	digest := sha256.Sum256(cert)*/
	return []byte{}, nil
	
}

