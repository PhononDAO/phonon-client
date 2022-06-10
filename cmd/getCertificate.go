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

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/config"
	"github.com/GridPlus/phonon-client/orchestrator"
	"github.com/spf13/cobra"
)

// getCertificateCmd represents the getCertificate command
var getCertificateCmd = &cobra.Command{
	Use:   "getCertificate",
	Short: "Gets the card certificate and validates it's signature with the configured CA",
	Long:  `Gets the card certificate and validates it's signature with the configured CA`,
	Run: func(cmd *cobra.Command, args []string) {
		getCertificate()
	},
}

func init() {
	rootCmd.AddCommand(getCertificateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCertificateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCertificateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getCertificate() {
	cs, err := card.QuickSecureConnection(readerIndex, false)
	if err != nil {
		fmt.Println("error establishing secure connection to card. err: ", err)
		return
	}
	s, err := orchestrator.NewSession(cs)
	if err != nil {
		fmt.Println("error creating new session. err: ", err)
		return
	}
	cardCert, err := s.GetCertificate()
	if err != nil {
		fmt.Println("error getting cert. err: ", err)
		return
	}
	conf, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error loading config. err: ", err)
		return
	}
	fmt.Println("Certificate: ", cardCert)
	//Validate cert matches configured CA
	err = cert.ValidateCardCertificate(*cardCert, conf.AppletCACert)
	if err != nil {
		fmt.Printf("error validating card certificate %v against CA certificate %v. err: %v\n", *cardCert, conf.AppletCACert, err)
		return
	}
	fmt.Println("getCertificate command successfully validated cert")
}
