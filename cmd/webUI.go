package cmd

import (
	"github.com/GridPlus/phonon-client/gui"
	"github.com/spf13/cobra"
)

// webUICmd represents the webUI command
var webUICmd = &cobra.Command{
	Use:   "webUI",
	Short: "Run the backend of the webui",
	Long: `Start a rest api to handle operations with the card.
	Meant to be paired with a graphical frontend. Not for production
	use as there is currently no security beyond the pin of the card.`,
	Run: func(cmd *cobra.Command, args []string) {
		gui.Server()
	},
}

func init() {
	rootCmd.AddCommand(webUICmd)
}
