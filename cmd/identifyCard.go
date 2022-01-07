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
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/GridPlus/phonon-client/card"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// identifyCardCmd represents the identifyCard command
var identifyCardCmd = &cobra.Command{
	Use:   "identifyCard",
	Short: "Request card identity information",
	Long: `Requests the card return it's identity public key along with a signature over
	a supplied nonce, proving it's possession of the correcsponding private key`,
	Run: func(cmd *cobra.Command, args []string) {
		identifyCard()
	},
}

func init() {
	rootCmd.AddCommand(identifyCardCmd)
}

func identifyCard() {
	cs, err := card.Connect(readerIndex)
	if err != nil {
		return
	}
	_, selectCardPubKey, _, err := cs.Select()
	if err != nil {
		fmt.Println("could not select applet during initialization:", err)
		return
	}
	fmt.Println("received pubkey from select:\n", hex.Dump(ethcrypto.FromECDSAPub(selectCardPubKey)))

	nonce := make([]byte, 32)
	rand.Read(nonce)
	cardPubKey, cardSig, err := cs.IdentifyCard(nonce)
	if err != nil {
		fmt.Println("error identifying card: ", err)
		return
	}

	log.Debug("cardPubKey:\n", hex.Dump(append(cardPubKey.X.Bytes(), cardPubKey.Y.Bytes()...)))
	log.Debug("cardSig:\n", hex.Dump(append(cardSig.R.Bytes(), cardSig.S.Bytes()...)))

	log.Debugf("ecdsaCardPubKey: % X\n", append(cardPubKey.X.Bytes(), cardPubKey.Y.Bytes()...))
	log.Debugf("ecdsaSignature: % X\n", append(cardSig.R.Bytes(), cardSig.S.Bytes()...))
	//Validate sig

	valid := ecdsa.Verify(cardPubKey, nonce, cardSig.R, cardSig.S)
	if !valid {
		log.Error("card signature on nonce not valid")
		return
	}
	log.Debug("identify card signature on nonce valid")
}
