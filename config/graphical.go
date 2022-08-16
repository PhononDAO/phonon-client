package config

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
)

var loggingTestURL = "https://logs.phonon.network/testKey"

func GraphicalConfiguration() {
	a := app.New()
	w := a.NewWindow("Configure Phonon Application")
	welcome := widget.NewLabel("Welcome to Phonon configuration. Please paste your Phonon Telemetry key here")
	keyBox := widget.NewEntry()
	w.SetContent(container.NewVBox(
		welcome,
		keyBox,
		widget.NewButton("Save Configuration", func() {
			cleaned := strings.Trim(keyBox.Text, `\ "`)
			err := CheckTelemetryKey(cleaned)
			if err != nil {
				welcome.SetText(fmt.Sprintf("Telemetry key validation failed: %s", err.Error()))
				return
			}
			SetDefaultConfig()
			viper.Set("TelemetryKey", cleaned)
			err = SaveConfig()
			if err != nil {
				welcome.SetText(fmt.Sprintf("Unable to save configuration: %s", err.Error()))
				return
			}
			welcome.SetText("Configuration saved. You may now exit the program")
		}),
		widget.NewButton("Stop seeing this message (will not save a telemetry key)", func() {
			SetDefaultConfig()
			err := SaveConfig()
			if err != nil {
				welcome.SetText(fmt.Sprintf("Unable to save default configuration: %s", err.Error()))
			} else {
				w.Close()
			}
		}),
		widget.NewButton("Skip for now/exit", func() {
			w.Close()
		}),
	))
	w.ShowAndRun()

}

func CheckTelemetryKey(key2check string) error {
	urlstruct, err := url.Parse(loggingTestURL)
	if err != nil {
		return err
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL:    urlstruct,
		Header: http.Header{
			"AuthToken": []string{key2check},
		},
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(respBytes))
	}
	return nil
}
