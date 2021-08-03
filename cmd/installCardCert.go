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
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/GridPlus/phonon-client/card"
	yubihsm "github.com/certusone/yubihsm-go"
	"github.com/certusone/yubihsm-go/commands"
	"github.com/certusone/yubihsm-go/connector"
	"github.com/ebfe/scard"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// installCardCert represents the installCardCert command
var installCardCert = &cobra.Command{
	Use:   "installCardCert",
	Short: "This is to sign and install an identity certificate to the card",
	Run: func(cmd *cobra.Command, args []string) {
		InstallCardCert() //todo rename this
	},
}

var (
	useDemoKey      bool
	yubikeySlot     string
	yubikeyPass     string
	usePhononApplet bool
)

func init() {
	rootCmd.AddCommand(installCardCert)

	installCardCert.Flags().BoolVarP(&useDemoKey, "demo", "d", false, "Use the demo key to sign -- insecure for demo purposes only")

	installCardCert.Flags().StringVarP(&yubikeySlot, "slot", "s", "", "Slot in which the signing ubikey is insterted") //this is taken in as a string to allow for a nil value instead of 0 value
	installCardCert.Flags().StringVarP(&yubikeyPass, "pass", "", "", "Ubikey Password")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCardCertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCardCertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func InstallCardCert() {
	var signKeyFunc func([]byte) ([]byte, error)
	// Determine Signing Key
	if useDemoKey {
		fmt.Println("Using Demo Key!")
		signKeyFunc = SignWithDemoKey
	} else {
		//gather information for ubikey signing
		var yubikeySlotInt int
		yubikeySlotInt, err := strconv.Atoi(yubikeySlot)
		if err != nil {
			fmt.Println("Please enter the yubikey slot: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			input = strings.TrimSuffix(input, "\n")
			yubikeySlotInt, err = strconv.Atoi(input)
		}
		if yubikeyPass == "" {
			fmt.Println("Please enter the yubikey password:")
			passBytes, err := terminal.ReadPassword(0)
			if err != nil {
				log.Fatalf("Unable to retrieve password from console: %s", err.Error())
			}
			yubikeyPass = string(passBytes)
		}

		signKeyFunc = SignWithYubikeyFunc(yubikeySlotInt, yubikeyPass)
	}

	// Select Card if multiple. Otherwise go with first one or error out
	cs, err := scConnectInteractive()
	if err != nil {
		log.Fatalf("Unable to connect to card: %s", err.Error())
	}

	err = cs.InstallCertificate(signKeyFunc)
	if err != nil {
		log.Fatalf("Unable to Install Certificate: %s", err.Error())
	}
}

func SignWithYubikeyFunc(slot int, password string) func([]byte) ([]byte, error) {
	return func(cert []byte) ([]byte, error) {

		c := connector.NewHTTPConnector("localhost:1234")
		sm, err := yubihsm.NewSessionManager(c, 1, "password")
		if err != nil {
			panic(err)
		}

		digest := sha256.Sum256(cert)
		uint16slot := uint16(slot)
		cmd, err := commands.CreateSignDataEcdsaCommand(uint16slot, digest[:])
		if err != nil {
			return []byte{}, err
		}
		res, err := sm.SendEncryptedCommand(cmd)
		if err != nil {
			return []byte{}, err
		}

		return res.(*commands.SignDataEddsaResponse).Signature, nil
	}
}

func SignWithDemoKey(cert []byte) ([]byte, error) {
	demoKey := []byte{
		0x03, 0x8D, 0x01, 0x08, 0x90, 0x00, 0x00, 0x00,
		0x10, 0xAA, 0x82, 0x07, 0x09, 0x80, 0x00, 0x00,
		0x01, 0xBB, 0x03, 0x06, 0x90, 0x08, 0x35, 0xF9,
		0x10, 0xCC, 0x04, 0x85, 0x09, 0x00, 0x00, 0x91,
	}
	var key ecdsa.PrivateKey

	// print("signing with demo key")
	key.D = new(big.Int).SetBytes(demoKey)
	key.PublicKey.Curve = secp256k1.S256()
	key.PublicKey.X, key.PublicKey.Y = key.PublicKey.Curve.ScalarBaseMult(key.D.Bytes())
	digest := sha256.Sum256(cert)
	ret, err := key.Sign(rand.Reader, digest[:], nil)
	if err != nil {
		return []byte{}, err
	}
	print("finished signing")
	return ret, nil

}

func scConnectInteractive() (*card.PhononCommandSet, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		log.Fatalf("Unable to establish Smart Card context: %s", err.Error())
		return nil, err
	}

	readers, err := ctx.ListReaders()
	if err != nil {
		log.Fatalf("Unable to list readers: %s", err.Error())
		return nil, err
	}
	if len(readers) == 0 {
		return nil, card.ErrReaderNotFound
	} else if len(readers) == 1 {
		return card.ConnectWithContext(ctx, 0)
	} else {
		fmt.Println("Please Select the index of the  card you wish to use:")
		for i, reader := range readers {
			fmt.Printf("%d: %s", i, string(reader))
		}
		reader := bufio.NewReader(os.Stdin)
		cardIndexStr, err := reader.ReadString('\n')
		cardIndexStr = strings.Trim(cardIndexStr, "\n")
		if err != nil {
			return nil, err
		}
		cardIndexInt, err := strconv.Atoi(cardIndexStr)
		if err != nil {
			return nil, err
		}
		return card.ConnectWithContext(ctx, cardIndexInt)
	}
}
