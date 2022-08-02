package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
)

func GraphicalConfiguration() {
	a := app.New()
	w := a.NewWindow("Configure Phonon Application")
	welcome := widget.NewLabel("Welcome to Phonon configuration. Please paste your alpha logging key here")
	keyBox := widget.NewEntry()
	w.SetContent(container.NewVBox(
		welcome,
		keyBox,
		widget.NewButton("Save Configuration", func() {
			cleaned := strings.Trim(keyBox.Text, `\ "`)
			SetDefaultConfig()
			viper.Set("TelemetryKey", cleaned)
			err := SaveConfig()
			if err != nil {
				welcome.SetText(fmt.Sprintf("Unable to save configuration: %s", err.Error()))
			}
			welcome.SetText("Configuration saved. You may now exit the program")
		}),
	))
	w.ShowAndRun()

}

func SaveConfig() error {
	viper.SetConfigType("yml")
	var configPath string
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		configPath = homedir + "/.phonon/"
		err = os.MkdirAll(configPath, 0700)
		if err != nil {
			return err
		}
	case "windows":
		configPath = homedir + "\\.phonon\\"
		os.MkdirAll(configPath, 0700)
	default:
		return fmt.Errorf("unable to set configuration path for %s", runtime.GOOS)
	}

	err = viper.WriteConfigAs(configPath + "phonon.yml")
	return err
}
