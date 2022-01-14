package config

import (
	"github.com/GridPlus/phonon-client/cert"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	PhononCommandSetConfig PhononCommandSetConfig
}

type PhononCommandSetConfig struct {
	PhononCACert []byte    //One of the CA certificates listed in the cert package
	LogLevel     log.Level //A logrus logLevel
}

func GetConfig() Config {
	//Add viper/commandline integration later
	conf := Config{
		PhononCommandSetConfig: PhononCommandSetConfig{
			PhononCACert: cert.PhononDemoCAPubKey,
			LogLevel:     log.DebugLevel,
		},
	}
	return conf
}
