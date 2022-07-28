package gui

import (
	"github.com/gotk3/gotk3/gtk"
	log "github.com/sirupsen/logrus"
)

func setupGUI(kill chan struct{}) {
	gtk.Init(nil)
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Phonon Testnet Config")
	win.Connect("destroy", func() {
	})
	l, err := gtk.LabelNew("Welcome to Phonon")
	if err != nil {
		kill <- struct{}{}
		log.Fatal("Unable to create label:", err)
	}
	win.Add(l)
	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk.Main()

}
