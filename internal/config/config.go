package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Env      string   `mapstructure:"env"`
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	Clerk    Clerk    `mapstructure:"clerk"`
}

type Server struct {
	Port string `mapstructure:"port"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type Clerk struct {
	PEMPublicKey      string   `mapstructure:"pem_public_key"`
	AuthorizedParties []string `mapstructure:"authorized_parties"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config") // For testing or different run contexts

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
		return nil, err
	}

	return &cfg, nil
}
