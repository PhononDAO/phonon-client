package config

import (
	"strings"

	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/hooks"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	//Global
	LogLevel log.Level //A logrus logLevel
	//PhononCommandSet
	AppletCACert []byte //One of the CA certificates listed in the cert package. Used as default if Certificate not set
	Certificate  string //string ID to select a certificate
	// logg exporting
	TelemetryKey string
}

func DefaultConfig() Config {
	//Add viper/commandline integration later
	conf := Config{
		Certificate:  "alpha",
		AppletCACert: cert.PhononAlphaCAPubKey,
		LogLevel:     log.DebugLevel,
	}
	return conf
}

func SetDefaultConfig() {
	viper.SetDefault("AppletCACert", cert.PhononAlphaCAPubKey)
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

	//Select cert based on provided certificate name
	if config.Certificate != "" {
		switch strings.ToLower(config.Certificate) {
		case "alpha", "testnet":
			break
		case "dev", "demo":
			config.AppletCACert = cert.PhononDemoCAPubKey
		default:
			break
		}
	}
	return config, nil
}
