// internal/config/config.go
package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort       string `mapstructure:"SERVER_PORT"`
	DBHost           string `mapstructure:"DB_HOST"`
	DBPort           string `mapstructure:"DB_PORT"`
	DBUser           string `mapstructure:"DB_USER"`
	DBPassword       string `mapstructure:"DB_PASSWORD"`
	DBName           string `mapstructure:"DB_NAME"`
	AccessJWTSecret  string `mapstructure:"ACCESS_JWT_SECRET"`
	RefreshJWTSecret string `mapstructure:"REFRESH_JWT_SECRET"`
}

func Load() (*Config, error) {
	// Set default values
	viper.SetDefault("SERVER_PORT", "8080")

	// Tell Viper to look for the .env file
	viper.SetConfigName(".env") // Name of config file (without extension)
	viper.SetConfigType("env")  // Type of config file
	viper.AddConfigPath(".")    // Look for config in the working directory
	viper.AddConfigPath("../")  // Look in parent directory too

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("Warning: .env file not found")
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %s", err)
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %v", err)
	}

	// Validate required fields
	if config.DBHost == "" || config.DBUser == "" || config.DBPassword == "" || config.DBName == "" {
		return nil, fmt.Errorf("required database configuration missing")
	}

	if config.AccessJWTSecret == "" {
		return nil, fmt.Errorf("ACCESS_JWT_SECRET is required")
	}
	
	if config.RefreshJWTSecret == "" {
		return nil, fmt.Errorf("ACCESS_JWT_SECRET is required")
	}

	return &config, nil
}
