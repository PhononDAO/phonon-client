package main

import (
	"github.com/GridPlus/phonon-client/internal/config"
	"github.com/GridPlus/phonon-client/internal/config/hooks"
	"github.com/GridPlus/phonon-client/pkg/gui"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Unable to load configuration")
	}

	if cfg.TelemetryKey != "" {
		log.Debug("setting up logging hook")
		log.AddHook(hooks.NewLoggingHook(cfg.TelemetryKey))
	}

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})
	log.Debug("starting local api server")

	// parse configuration
	//todo: make a graphical window pop up indicating an error state
	//////////////////////////////////
	//     An error has occured:    //
	//     "something something"    //
	//        --------              //
	//        |  ok  |              //
	//        ________              //
	//////////////////////////////////

	// initialize backends

	gui.Server("8080", "", "", false, log.StandardLogger(), cfg.Certificate)

}
