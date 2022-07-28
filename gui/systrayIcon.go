package gui

import (
	"fyne.io/systray"
	log "github.com/sirupsen/logrus"
)

func SystrayIcon(kill chan struct{}) (end func()) {
	startSystray, endSystray := systray.RunWithExternalLoop(onReady, onExit(kill))
	startSystray()
	return endSystray
}

func onReady() {
	systray.SetIcon(phononLogo)
	systray.SetTitle("")
	systray.SetTooltip("Phonon UI backend is currently running")
	mQuit := systray.AddMenuItem("Quit", "Exit PhononUI")
	mQuit.SetIcon(xIcon)
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit(kill chan struct{}) func() {
	return func() {
		log.Println("Server killed by systray interaction")
		kill <- struct{}{}
	}
}