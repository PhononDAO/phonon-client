package main

import (
	"os"

	"github.com/GridPlus/phonon-client/card"
	log "github.com/sirupsen/logrus"
)

func main() {
	cs, err := card.Connect()
	if err != nil {
		log.Error("could not connect to card. err: ", err)
		os.Exit(1)
	}
	err = cs.Select()
	if err != nil {
		log.Error("error selecting phonon applet. err: ", err)
	}
}
