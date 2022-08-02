/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/GridPlus/phonon-client/config"
	"github.com/spf13/cobra"
)

// graphicalConfigureCmd represents the graphicalConfigure command
var graphicalConfigureCmd = &cobra.Command{
	Use:   "graphicalConfigure",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(_ *cobra.Command, _ []string) {
		config.GraphicalConfiguration()
	},
}

func init() {
	rootCmd.AddCommand(graphicalConfigureCmd)

}
