package cfg

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represent application configuration
type Config struct {
	BotToken  string
	DBConn    string
	ImagesDir string
}

func GetConfig() *Config {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	viper.SetConfigType("toml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		logrus.WithError(err).Fatal("Error in config file")
	}

	config := &Config{}
	err = viper.Unmarshal(config)
	if err != nil {
		logrus.WithError(err).Panic("Failed to unmarshal config")
	}

	return config
}
