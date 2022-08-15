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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/GridPlus/phonon-client/card"
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/usb"
	"github.com/GridPlus/phonon-client/util"

	"github.com/spf13/cobra"
)

// provisionCardsCmd represents the provisionCards command
var provisionCardsCmd = &cobra.Command{
	Use:   "provisionCards",
	Short: "Flash and cert phonon cards",
	Long: `Provision phonon cards in parallel. Flash new or previously installed phonon cards,
	install the proper signed certificate, and if necessary register them with the testnet rewards address service. `,
	Run: func(cmd *cobra.Command, args []string) {
		provisionCards()
	},
}

type provisioningReport struct {
	readerIdentifier string
	cardID           string
	completionTime   time.Time
	err              error
}

var (
	java8BinPath       string
	gpJarPath          string
	capFilePath        string
	ISDUnlockCode      string
	registrationAPIKey string
	rollCallMode       bool
	noRegistration     bool
	//using yubikey vars from installCardCert
)

func init() {
	rootCmd.AddCommand(provisionCardsCmd)

	provisionCardsCmd.Flags().StringVarP(&java8BinPath, "java", "j", "/Library/Java/JavaVirtualMachines/jdk1.8.0_301.jdk/Contents/Home/bin/java", "Path to java 8 bin executable")
	provisionCardsCmd.Flags().StringVarP(&gpJarPath, "gp", "g", "../bin/globalplatformpro/gp.jar", "Path to globalplatform gp.jar file, typically from Phonon-Build repo")
	provisionCardsCmd.Flags().StringVarP(&capFilePath, "cap", "c", "", "path to CAP file to install")
	provisionCardsCmd.Flags().StringVarP(&ISDUnlockCode, "isd", "d", "", "Manufacturer ISD Unlock Code. Must be specific to variety of card being flashed")
	provisionCardsCmd.Flags().BoolVarP(&noRegistration, "no-registration", "n", false, "Enable to skip registering flashed cards with the phonon address service. Registration enabled by default")
	provisionCardsCmd.Flags().StringVarP(&registrationAPIKey, "api-key", "a", "", "Api-key for adding fresh cards to registration service")
	provisionCardsCmd.Flags().BoolVarP(&rollCallMode, "roll-call", "r", false, "Enters roll call mode, which lights up readers one by one so they can be physically identified before the flashing run. Reader names stay consistent when the program is restarted for flashing, but it is untested whether these remain consistent over OS restarts, or when attaching and detaching readers from USB.")
	provisionCardsCmd.Flags().IntVarP(&yubikeySlot, "yubi-slot", "s", 0, "Slot in which the signing yubikey is insterted") //this is taken in as a string to allow for a nil value instead of 0 value
	provisionCardsCmd.Flags().StringVarP(&yubikeyPass, "yubi-pass", "p", "", "Yubikey Password")

	provisionCardsCmd.MarkFlagRequired("cap")
	provisionCardsCmd.MarkFlagRequired("isd")
	provisionCardsCmd.MarkFlagRequired("yubi-slot")
	provisionCardsCmd.MarkFlagRequired("yubi-pass")
}

func provisionCards() {
	startTime := time.Now()
	cards, err := usb.ConnectAllUSBReaders()
	if err != nil {
		fmt.Println("error connecting all readers: ", err)
		return
	}

	wg := new(sync.WaitGroup)

	reports := make(chan provisioningReport, len(cards))

	for i, sc := range cards {
		cardStatus, err := sc.Status()
		if err != nil {
			fmt.Println("error getting card status: ", err)
		}
		readerName := cardStatus.Reader
		fmt.Println("reader index: ", i)
		fmt.Println("reader name: ", readerName)

		if rollCallMode {
			rollCall(readerName, i)
		} else {
			wg.Add(1)
			go provisionCardAsync(i, readerName, wg, reports)
		}
	}

	//If just calling roll exit before Wait as waitgroup is not being utilized
	if rollCallMode {
		fmt.Println("roll call for all readers complete")
		return
	}

	//Wait until all cards complete provisioning to terminate
	wg.Wait()

	//Loop through all results and print status in tab aligned table
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "status\ttime elapsed\treader ID\tcardID\terror (if any)\t")
	for {
		select {
		case report := <-reports:
			var result string
			if report.err == nil {
				result = "success"
			} else {
				result = "failure"
			}
			timeElapsed := report.completionTime.Sub(startTime)
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t\n", result, timeElapsed, report.readerIdentifier, report.cardID, report.err)

		default:
			w.Flush()
			fmt.Println("all cards completed provisioning attempt")
			fmt.Printf("run started: %v| run finished %v\n", startTime.Local().Format(time.RFC822), time.Now().Local().Format(time.RFC822))
			fmt.Println("total time elapsed: ", time.Since(startTime))
			return
		}
	}
}

func provisionCardAsync(readerIndex int, readerName string, wg *sync.WaitGroup, reportsChan chan provisioningReport) {
	defer wg.Done()
	cardID, err := provisionCard(readerIndex, readerName)
	report := provisioningReport{
		readerIdentifier: readerName + " : " + fmt.Sprint(readerIndex),
		cardID:           cardID,
		completionTime:   time.Now(),
		err:              err,
	}
	reportsChan <- report
}

func provisionCard(readerIndex int, readerName string) (cardID string, err error) {
	//ISD Unlock card?
	unlockCmd := exec.Command("opensc-tool",
		"-r", readerName,
		"-s", ISDUnlockCode)
	output, err := unlockCmd.Output()
	if err != nil {
		fmt.Printf("error unlocking card on reader %v: %v\n", readerName, err)
	}
	fmt.Println(string(output))
	//Delete any existing Phonon App from previous install
	//error is expected on brand new cards, so check logs but continues on
	deleteCmd := exec.Command(java8BinPath,
		"-jar", gpJarPath,
		"--reader", readerName,
		"--delete", "A0000008200003")
	output, err = deleteCmd.Output()
	if err != nil {
		fmt.Printf("error deleting applet on reader %v: %v\n", readerName, err)
		fmt.Println("proceeding on assumption applet hasn't been installed before")
	}
	fmt.Println(string(output))

	//Install CAP file to card
	output, err = exec.Command(java8BinPath,
		"-jar", gpJarPath,
		"--reader", readerName,
		"--install", capFilePath,
		"--applet", "A0000008200003",
		"--package", "A0000008200003").Output()
	if err != nil {
		fmt.Printf("error installing applet on reader: %v: %v\n", readerName, err)
		return "", err
	}
	fmt.Println(string(output))
	fmt.Printf("applet successfully installed on reader: %v\n", readerName)

	cs, err := card.Connect(readerIndex)
	if err != nil {
		fmt.Printf("unable to connect to newly flashed applet on reader %v: %v\n", readerName, err)
		return "", err
	}

	_, _, _, err = cs.Select()
	if err != nil {
		fmt.Printf("unable to select applet on reader %v: %v\n", readerName, err)
		return "", err
	}

	err = cs.InstallCertificate(cert.SignWithYubikeyFunc(yubikeySlot, yubikeyPass))
	if err != nil {
		fmt.Printf("error installing certificate on reader %v: %v\n", readerName, err)
		return "", err
	}

	//Validate card installed successfully
	pubkey, _, err := cs.IdentifyCard(util.RandomKey(32))
	if err != nil {
		fmt.Printf("validation via IdentifyCard failed on reader %v: %v\n", readerName, err)
		return "", err
	}
	if !noRegistration {
		cardID = util.CardIDFromPubKey(pubkey)
		output, err := registerCard(cardID)
		if err != nil {
			fmt.Printf("error registering card %v on reader %v. err: %v\n", cardID, readerName, err)
			return cardID, err
		}
		fmt.Printf("card %v registered on reader: %v\n", cardID, readerName)
		fmt.Println("address-service response: ", output)
	}
	fmt.Println("card finished provisioning on reader: ", readerName)

	return cardID, nil
}

//registerCard takes a cardID and registers it with the phonon-address-service for use in the testnet
func registerCard(cardID string) (string, error) {
	//TODO: centralize hardcoded values
	registrationSrvURI := "https://register.phonon.network/add"
	type AddRequest struct {
		ID string `json:"id"`
	}

	reqBody, err := json.Marshal(AddRequest{ID: cardID})
	if err != nil {
		fmt.Println("error marshaling registration /add request. err: ", err)
		return "", err
	}

	req, err := http.NewRequest("POST", registrationSrvURI, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("error forming registration request: ", err)
		return "", err
	}

	req.Header.Set("api-key", registrationAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

//Lights up readers one at a time by looping through the -info gp.jar command.
//Allows the operator to position the readers sequentially for easy identification of errors during runs
//as readers may switch IDs each time they power off and on.
func rollCall(readerName string, readerIndex int) {
	quit := make(chan bool)
	go func() {
		fmt.Println("lighting up", readerName)
		for {
			select {
			case <-quit:
				return
			default:
				//Get reader info to trigger blinking light
				//ignore errors
				exec.Command(java8BinPath,
					"-jar", gpJarPath,
					"--reader", readerName,
					"-info").Output()
			}
		}
	}()

	fmt.Println("press enter to move on to next reader")
	fmt.Scanln()
	quit <- true
}
