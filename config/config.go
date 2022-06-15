package config

import (
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/hooks"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	//Global
	LogLevel log.Level //A logrus logLevel
	//PhononCommandSet
	AppletCACert []byte //One of the CA certificates listed in the cert package

	//EthChainService
	EthChainServiceApiKey string
	EthNodeURL            string
	// logg exporting
	TelemetryKey string
}

type EthChainServiceConfig struct {
	ApiKey string
}

func DefaultConfig() Config {
	//Add viper/commandline integration later
	conf := Config{
		AppletCACert: cert.PhononDemoCAPubKey,
		LogLevel:     log.DebugLevel,
	}
	return conf
}

func SetDefaultConfig() {
	viper.SetDefault("AppletCACert", cert.PhononDemoCAPubKey)
	viper.SetDefault("LogLevel", log.DebugLevel)
}

func LoadConfig() (config Config, err error) {
	SetDefaultConfig()
	viper.AddConfigPath("$HOME/.phonon/")
	viper.AddConfigPath("$XDG_CONFIG_HOME/.phonon/phonon.yml")
	viper.AddConfigPath("/usr/var/phonon/phonon.yml")
	viper.SetConfigName("phonon")
	viper.SetConfigType("yml")

	viper.SetEnvPrefix("phonon")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		log.Debug("config file not found, using default config")
		return DefaultConfig(), nil
	}
	if err != nil {
		log.Error("unable to read configuration file, using default config. err: ", err)
		return DefaultConfig(), err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return DefaultConfig(), err
	}
	// Possibly not the best place to put this, but it does a good job of setting this up before an interactive session
	if config.TelemetryKey != "" {
		log.Debug("setting up logging hook")
		log.AddHook(hooks.NewLoggingHook(config.TelemetryKey))
	}

	return config, nil
}
