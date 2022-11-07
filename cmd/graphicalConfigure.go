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
	Short: "Open the graphical configuration window",
	Long:  `Opens a graphical configuration window that places a configuration file in the default location (platform dependent)`,
	Run: func(_ *cobra.Command, _ []string) {
		config.GraphicalConfiguration()
	},
}

func init() {
	rootCmd.AddCommand(graphicalConfigureCmd)

}
