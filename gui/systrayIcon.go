package gui

import (
	"fmt"

	"fyne.io/systray"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

//go:embed icons/phonon.ico
var phononLogo []byte

//go:embed icons/x.ico
var xIcon []byte

func SystrayIcon(kill chan struct{}, port string) (end func()) {
	startSystray, endSystray := systray.RunWithExternalLoop(onReadyFunc(port), onExit(kill))
	startSystray()
	fmt.Println("systray started")
	return endSystray
}

func onReadyFunc(port string) func() {
	return func() {
		fmt.Println("onready")
		systray.SetIcon(phononLogo)
		systray.SetTitle("")
		systray.SetTooltip("Phonon UI backend is currently running")
		mOpen := systray.AddMenuItem("Open Phonon UI", "Open the phonon ui in your browser")
		mOpen.SetIcon(phononLogo)
		mQuit := systray.AddMenuItem("Quit", "Exit PhononUI")
		mQuit.SetIcon(xIcon)
		go func() {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-mOpen.ClickedCh:
				browser.OpenURL("http://localhost:" + port + "/")
			}
		}()
	}
}

func onExit(kill chan struct{}) func() {
	return func() {
		log.Println("Server killed by systray interaction")
		kill <- struct{}{}
	}
}
