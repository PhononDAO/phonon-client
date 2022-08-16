package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/GridPlus/phonon-client/hooks"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	//Global
	LogLevel log.Level //A logrus logLevel
	//PhononCommandSet
	Certificate string //string ID to select a certificate
	// log exporting
	TelemetryKey string
}

func DefaultConfig() Config {
	//Add viper/commandline integration later
	conf := Config{
		Certificate: "alpha",
		LogLevel:    log.DebugLevel,
	}
	return conf
}

func SetDefaultConfig() {
	viper.SetDefault("LogLevel", log.DebugLevel)
}

func LoadConfig() (config Config, err error) {
	SetDefaultConfig()
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
