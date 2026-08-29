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
	Supabase  Supabase  `mapstructure:"supabase"`
	AWS       AWS       `mapstructure:"aws"`
}

// AWS holds AWS S3 configuration used for video storage.
type AWS struct {
	Region                   string `mapstructure:"region"`
	Bucket                   string `mapstructure:"bucket"`
	AccessKey                string `mapstructure:"access_key"`
	SecretKey                string `mapstructure:"secret_key"`
	CloudFrontDomain         string `mapstructure:"cf_domain"`
	CloudFrontKeyID          string `mapstructure:"cf_key_id"`
	CloudFrontPrivateKeyPath string `mapstructure:"cf_private_key_path"`
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
	Port           string   `mapstructure:"port"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
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

	// Unset placeholders so Viper falls back to config files like local.yaml
	if os.Getenv("SUPABASE_ANON_KEY") == "your_supabase_anon_key_here" {
		os.Unsetenv("SUPABASE_ANON_KEY")
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

	// Parse ALLOWED_ORIGINS environment variable if set (comma-separated list)
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		origins := strings.Split(envOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		cfg.Server.AllowedOrigins = origins
	}

	// Rate limit defaults
	if cfg.RateLimit.RPS == 0 {
		cfg.RateLimit.RPS = 10 // 10 requests per second
	}
	if cfg.RateLimit.Burst == 0 {
		cfg.RateLimit.Burst = 20 // burst of 20
	}

	// Propagate config properties back to environment variables so that standard library os.Getenv calls get the resolved values
	if cfg.Supabase.URL != "" {
		os.Setenv("SUPABASE_URL", cfg.Supabase.URL)
	}
	if cfg.Supabase.Key != "" {
		os.Setenv("SUPABASE_ANON_KEY", cfg.Supabase.Key)
	}
	if cfg.Supabase.ServiceRoleKey != "" {
		os.Setenv("SUPABASE_SERVICE_ROLE_KEY", cfg.Supabase.ServiceRoleKey)
	}
	if cfg.Supabase.Bucket != "" {
		os.Setenv("SUPABASE_BUCKET", cfg.Supabase.Bucket)
	}

	// Propagate AWS S3 config back to environment variables
	if cfg.AWS.Region != "" {
		os.Setenv("AWS_REGION", cfg.AWS.Region)
	}
	if cfg.AWS.Bucket != "" {
		os.Setenv("AWS_S3_BUCKET", cfg.AWS.Bucket)
	}
	if cfg.AWS.AccessKey != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", cfg.AWS.AccessKey)
	}
	if cfg.AWS.SecretKey != "" {
		os.Setenv("AWS_SECRET_ACCESS_KEY", cfg.AWS.SecretKey)
	}
	if cfg.AWS.CloudFrontDomain != "" {
		os.Setenv("AWS_CF_DOMAIN", cfg.AWS.CloudFrontDomain)
	}
	if cfg.AWS.CloudFrontKeyID != "" {
		os.Setenv("AWS_CF_KEY_ID", cfg.AWS.CloudFrontKeyID)
	}
	if cfg.AWS.CloudFrontPrivateKeyPath != "" {
		os.Setenv("AWS_CF_PRIVATE_KEY_PATH", cfg.AWS.CloudFrontPrivateKeyPath)
	}

	return &cfg, nil
}

type Supabase struct {
	URL            string `mapstructure:"url"`
	Key            string `mapstructure:"anon_key"`
	ServiceRoleKey string `mapstructure:"service_role_key"`
	Bucket         string `mapstructure:"bucket"`
}
