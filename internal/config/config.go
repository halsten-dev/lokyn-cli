package config

import (
	"errors"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

const (
	LANGUAGE      string = "language"
	DEEPL_API_KEY string = "deeplapikey"
)

func Init() {
	var configPath string
	var err error

	cfgDir, err := os.UserConfigDir()

	if err != nil {
		log.Panic("Failed to get user config dir")
	}

	configPath = filepath.Join(cfgDir, "lokyn_tui")

	err = os.MkdirAll(configPath, os.ModePerm)

	if err != nil {
		log.Panic("Failed to create the config directory : ", err)
	}

	// Set defaults
	viper.SetDefault(LANGUAGE, "en")
	viper.SetDefault(DEEPL_API_KEY, "")

	viper.SetConfigName("lokyn")
	viper.SetConfigType("toml")

	viper.AddConfigPath(filepath.Join(cfgDir, "lokyn_tui"))

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			err = viper.SafeWriteConfig()

			if err != nil {
				log.Panic("error while writing the config file : ", err)
			}
		} else {
			log.Panic("error while reading the config file : ", err)
		}
	}
}
