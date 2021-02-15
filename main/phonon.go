package main

import (
	"github.com/GridPlus/phonon-client/cmd"
	log "github.com/sirupsen/logrus"
)

type Phonon struct {
}

func main() {
	log.SetLevel(log.DebugLevel)

	cmd.Execute()
}
