package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env       string    `mapstructure:"env"`
	Server    Server    `mapstructure:"server"`
	Database  Database  `mapstructure:"database"`
	Clerk     Clerk     `mapstructure:"clerk"`
	GitHub    GitHub    `mapstructure:"github"`
	Gemini    Gemini    `mapstructure:"gemini"`
	RateLimit RateLimit `mapstructure:"rate_limit"`
}

// RateLimit controls the global per-IP rate limiter.
// Defaults: 10 requests/sec, burst of 20.
type RateLimit struct {
	RPS   float64 `mapstructure:"rps"`   // sustained requests per second
	Burst int     `mapstructure:"burst"` // max burst size
}

type Gemini struct {
	APIKey string `mapstructure:"api_key"`
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
	PEMPublicKey         string   `mapstructure:"pem_public_key"`
	AuthorizedParties    []string `mapstructure:"authorized_parties"`
	WebhookSigningSecret string   `mapstructure:"webhook_signing_secret"`
}

type GitHub struct {
	PAT string `mapstructure:"github_pat"`
}

func LoadConfig() (*Config, error) {
	// Load .env file only in non-production environments
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load() // Ignore error if .env doesn't exist
	}

	viper.SetConfigName("local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")

	// 🔑 CRITICAL: Map nested keys to env vars (database.host → DATABASE_HOST)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file only if present (not required in production)
	if err := viper.ReadInConfig(); err != nil {
		log.Println("No local config file found, using env vars only")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
		return nil, err
	}

	// Validate required fields
	if cfg.Database.Host == "" {
		log.Fatal("DATABASE_HOST is required")
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080" // Default port
	}

	// Rate limit defaults
	if cfg.RateLimit.RPS == 0 {
		cfg.RateLimit.RPS = 10 // 10 requests per second
	}
	if cfg.RateLimit.Burst == 0 {
		cfg.RateLimit.Burst = 20 // burst of 20
	}

	return &cfg, nil
}
