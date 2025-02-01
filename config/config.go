package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type EmailConfig struct {
	SMTPHost     string `mapstructure:"SMTP_HOST" mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"SMTP_PORT" mapstructure:"smtp_port"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME" mapstructure:"smtp_username"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD" mapstructure:"smtp_password"`
	FromEmail    string `mapstructure:"FROM_EMAIL" mapstructure:"from_email"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type Config struct {
	Env              string `mapstructure:"ENV"`
	Component        string `mapstructure:"COMPONENT"`
	ServerPort       string `mapstructure:"SERVER_PORT"`
	DBHost           string `mapstructure:"DB_HOST"`
	DBPort           string `mapstructure:"DB_PORT"`
	DBUser           string `mapstructure:"DB_USER"`
	DBPassword       string `mapstructure:"DB_PASSWORD"`
	DBName           string `mapstructure:"DB_NAME"`
	AccessJWTSecret  string `mapstructure:"ACCESS_JWT_SECRET"`
	RefreshJWTSecret string `mapstructure:"REFRESH_JWT_SECRET"`
	EmailConfig      `mapstructure:",squash"`
	RedisConfig      `mapstructure:",squash"`
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

	// Ensure Viper reads all environment variables correctly
	viper.BindEnv("SMTP_HOST")
	viper.BindEnv("SMTP_PORT")
	viper.BindEnv("SMTP_USERNAME")
	viper.BindEnv("SMTP_PASSWORD")
	viper.BindEnv("FROM_EMAIL")
	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("ACCESS_JWT_SECRET")
	viper.BindEnv("REFRESH_JWT_SECRET")

	// Unmarshal into the Config struct
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

	// Make sure email config is set
	if config.EmailConfig.SMTPHost == "" || config.EmailConfig.SMTPPort == 0 || config.EmailConfig.SMTPUsername == "" || config.EmailConfig.SMTPPassword == "" || config.EmailConfig.FromEmail == "" {
		return nil, fmt.Errorf("required email configuration missing")
	}

	return &config, nil
}
