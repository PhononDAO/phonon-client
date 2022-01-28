package cmd

import (
	"github.com/GridPlus/phonon-client/gui"
	"github.com/spf13/cobra"
)

var (
	guiCert, guiKey, guiPort string
	useMock                  bool
)

// webUICmd represents the webUI command
var webUICmd = &cobra.Command{
	Use:   "webUI",
	Short: "Run the backend of the webui",
	Long: `Start a rest api to handle operations with the card.
	Meant to be paired with a graphical frontend. Not for production
	use as there is currently no security beyond the pin of the card.`,
	Run: func(cmd *cobra.Command, args []string) {
		gui.Server(guiPort, guiCert, guiKey, useMock)
	},
}

func init() {
	rootCmd.AddCommand(webUICmd)
	webUICmd.Flags().StringVarP(&guiPort, "port", "p", "8080", "port for clients to connect on")
	webUICmd.Flags().StringVarP(&guiCert, "cert", "c", "", "SSL certificate")
	webUICmd.Flags().StringVarP(&guiKey, "key", "k", "", "SSL key")
	webUICmd.Flags().BoolVarP(&useMock, "useMock", "m", false, "generate a mock card for testing")
}
