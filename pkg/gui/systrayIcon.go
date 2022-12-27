package gui

import (
	"fmt"
	"runtime"

	_ "embed"

	"fyne.io/systray"
	"github.com/pkg/browser"
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

func SystrayIcon(port string) {
	systray.Run(onReadyFunc(port), onExit)
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
		systray.SetTitle("Phonon")
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
		fmt.Println("systray started")
	}
}

func onExit() {
}
