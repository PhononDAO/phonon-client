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
	"github.com/spf13/cobra"
	"time"
)

// mineNativePhononsCmd represents the mineNativePhonons command
var mineNativePhononsCmd = &cobra.Command{
	Use:   "mineNativePhonons [duration]",
	Short: "Begin mining native phonons",
	Long: `Begin mining native phonons.
	If called with no arguments command will repeatedly mine for phonons until cancelled.
	Pass a duration in go time syntax to mine for a specific duration instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		mineNativePhonons()
	},
}

var difficulty uint8
var hashMining bool

func init() {
	rootCmd.AddCommand(mineNativePhononsCmd)
	mineNativePhononsCmd.PersistentFlags().Uint8VarP(&difficulty, "difficulty", "d", 1, "express the desired difficulty of the mining operation, in bytes with leading zeros")
}

func mineNativePhonons() {
	//Connect and Pair with Card
	cs, err := card.QuickSecureConnection(readerIndex, staticPairing)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = cs.VerifyPIN("111111")
	if err != nil {
		fmt.Println("unable to verify PIN: ", err)
		return
	}
	var totalTime time.Duration
	var i int
	for i = 1; i > 0; i++ {
		fmt.Println("mining attempt #", i)
		start := time.Now()
		keyIndex, hash, err := cs.MineNativePhonon(difficulty)
		elapsed := time.Since(start)
		totalTime += elapsed
		fmt.Println("mining iteration duration: ", elapsed)
		if err == card.ErrMiningFailed {
			fmt.Println("mining failed to find phonon. repeating attempt...")
		} else if err != nil {
			fmt.Println("unknown error mining phonon. err: ", err)
			return
		} else {
			fmt.Printf("mined native phonon after %v attempts.\n", i)
			fmt.Printf("keyIndex: %v\nhash: % X\n", keyIndex, hash)
			break
		}
	}
	fmt.Printf("\nmining completed with difficulty of %v bit(s): \n", difficulty)
	fmt.Println("total mining attempts: ", i)
	fmt.Println("mining run elapsed time: ", totalTime)
	averageTime := time.Duration(float64(totalTime.Nanoseconds()) / float64(i))

	fmt.Println("average iteration time: ", averageTime)
}
