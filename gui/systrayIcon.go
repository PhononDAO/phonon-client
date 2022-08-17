package gui

import (
	"fmt"
	"runtime"

	_ "embed"

	"fyne.io/systray"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

//go:embed icons/phonon.png
var phononLogoPng []byte

//go:embed icons/x.png
var xIconPng []byte

//go:embed icons/phonon.ico
var phononLogoIco []byte

//go:embed icons/x.ico
var xIconIco []byte

var phononLogo []byte
var xIcon []byte

func SystrayIcon(kill chan struct{}, port string) (end func()) {
	startSystray, endSystray := systray.RunWithExternalLoop(onReadyFunc(port), onExit(kill))
	startSystray()
	fmt.Println("systray started")
	return endSystray
}

func onReadyFunc(port string) func() {
	switch runtime.GOOS {
	case "linux", "darwin":
		phononLogo = phononLogoPng
		xIcon = xIconPng
	case "windows":
		phononLogo = phononLogoIco
		xIcon = xIconIco
	default:
		return func() {
			fmt.Println("Unsupported os")
			systray.Quit()
		}
	}
	return func() {
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
