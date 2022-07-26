/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

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

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/spf13/cobra"
)

// generateCAKeypairCmd represents the generateCAKeypair command
var generateCAKeypairCmd = &cobra.Command{
	Use:   "generateCAKeypair",
	Short: "Generates a keypair for use as a phonon card CA",
	Long: `Generates a keypair for us as a phonon card certificate authority.
	Prints the public and private key details in the string formats needed for inclusion
	in the phonon-card and phonon-client source code. `,
	Run: func(_ *cobra.Command, _ []string) {
		generateCAKeypair()
	},
}

func init() {
	rootCmd.AddCommand(generateCAKeypairCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCAKeypairCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCAKeypairCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func generateCAKeypair() {
	privKey, err := ethcrypto.GenerateKey()
	if err != nil {
		fmt.Println("error generating key. err: ", err)
	}
	//Print privKey as bytes
	fmt.Println("PrivKey:")
	fmt.Println(printGolangfmt(ethcrypto.FromECDSA(privKey)))
	//Print pubKey as golang formatted CA hex byte string
	fmt.Println("PubKey: ")
	fmt.Println(printGolangfmt(ethcrypto.FromECDSAPub(&privKey.PublicKey)))

	//Print pubKey as javacard formatted CA hex byte string
	fmt.Println("Javacard PubKey: ")
	fmt.Println(printJavacardfmt(ethcrypto.FromECDSAPub(&privKey.PublicKey)))
}

func printGolangfmt(key []byte) string {
	var result string
	width := 0
	for i, b := range key {
		if width%8 == 0 && width != 0 {
			result += "\n"
			width = 0
		}
		result += fmt.Sprintf("0x%02X, ", b)
		width += 1
		//Add newline after pubkey prefix byte
		if i == 0 && b == 0x04 {
			result += "\n"
			width = 0
		}

	}
	return result
}

//Same as golang fmt code except we prepend the string (byte) before any value over 0x80
func printJavacardfmt(key []byte) string {
	var result string
	width := 0
	for i, b := range key {
		if width%8 == 0 && width != 0 {
			result += "\n"
			width = 0
		}
		if b >= 0x80 {
			result += "(byte) "
		}
		result += fmt.Sprintf("0x%02X, ", b)
		width += 1
		//Add newline after pubkey prefix byte
		if i == 0 && b == 0x04 {
			result += "\n"
			width = 0
		}
	}
	return result
}
