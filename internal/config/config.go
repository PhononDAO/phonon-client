package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/GridPlus/phonon-core/pkg/cert"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigFile struct {
	//PhononCommandSet
	Certificate string //string ID to select a certificate
	// log exporting
	TelemetryKey string
	LoggingLevel string
}

type Config struct {
	TelemetryKey string
	Certificate  []byte
	Level        logrus.Level
}

func DefaultConfig() Config {
	//Add viper/commandline integration later
	conf := Config{
		Certificate: cert.PhononDemoCAPubKey,
	}
	return conf
}

func LoadConfig() (config Config, err error) {
	// SetDefaultConfig()
	switch runtime.GOOS {
	case "linux", "darwin":
		viper.AddConfigPath("$HOME/.phonon/")
		viper.AddConfigPath("$XDG_CONFIG_HOME/.phonon/phonon.yml")
		viper.AddConfigPath("/usr/var/phonon/phonon.yml")

	case "windows":
		viper.AddConfigPath("$HOME\\.phonon\\")
	default:
		return Config{}, fmt.Errorf("unknown os: %s encountered", runtime.GOOS)
	}
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

	var configFile ConfigFile

	err = viper.Unmarshal(&configFile)
	if err != nil {
		log.Error("error unmarshalling into config: ", err)
		return DefaultConfig(), err
	}
	switch strings.ToLower(configFile.Certificate) {
	case "demo":
		config.Certificate = cert.PhononDemoCAPubKey
	case "alpha":
		config.Certificate = cert.PhononAlphaCAPubKey
	case "mock":
		config.Certificate = cert.PhononMockCAPubKey

	default:
		return Config{}, fmt.Errorf("unknon option for CA key")
	}

	config.TelemetryKey = configFile.TelemetryKey

	if configFile.LoggingLevel == "" {
		config.Level = logrus.ErrorLevel
	} else {
		config.Level, err = logrus.ParseLevel(configFile.LoggingLevel)
		if err != nil {
			return Config{}, fmt.Errorf("unable to determine logging level from %s: %s", configFile.LoggingLevel, err.Error())
		}
	}

	return
}

func DefaultConfigPath() (string, error) {
	var ret string
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		ret = homedir + "/.phonon/"
		if err != nil {
			return "", err
		}
	case "windows":
		ret = homedir + "\\.phonon\\"
	default:
		return "", fmt.Errorf("unable to set configuration path for %s", runtime.GOOS)
	}
	return ret, nil
}

func SaveConfig() error {
	viper.SetConfigType("yml")
	configPath, err := DefaultConfigPath()
	if err != nil {
		return err
	}
	err = os.MkdirAll(configPath, 0700)
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(configPath + "phonon.yml")
}
