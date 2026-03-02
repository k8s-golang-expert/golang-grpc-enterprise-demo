package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv       string `mapstructure:"APP_ENV"`
	GRPCPort     string `mapstructure:"GRPC_PORT"`
	HTTPPort     string `mapstructure:"HTTP_PORT"`
	DatabaseURL  string `mapstructure:"DATABASE_URL"`
	DBHost       string `mapstructure:"DB_HOST"`
	DBPort       string `mapstructure:"DB_PORT"`
	DBUser       string `mapstructure:"DB_USER"`
	DBPassword   string `mapstructure:"DB_PASSWORD"`
	DBName       string `mapstructure:"DB_NAME"`
	DBSSLMode    string `mapstructure:"DB_SSLMODE"`
	JWTSecret    string `mapstructure:"JWT_SECRET"`
	JWTExpiryH   int    `mapstructure:"JWT_EXPIRY_HOURS"`
	RateLimitRPM int    `mapstructure:"RATE_LIMIT_RPM"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Explicitly bind ALL env vars so Unmarshal picks them up
	for _, key := range []string{
		"APP_ENV", "GRPC_PORT", "HTTP_PORT", "PORT",
		"DATABASE_URL",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"JWT_SECRET", "JWT_EXPIRY_HOURS", "RATE_LIMIT_RPM",
	} {
		_ = viper.BindEnv(key)
	}

	// defaults
	viper.SetDefault("GRPC_PORT", "50051")
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("RATE_LIMIT_RPM", 100)

	_ = viper.ReadInConfig() // ok if .env missing — env vars suffice

	// Railway sets PORT — use it as HTTP_PORT if HTTP_PORT not explicitly set
	if port := viper.GetString("PORT"); port != "" {
		viper.Set("HTTP_PORT", port)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Double-check: os.Getenv as final fallback for DATABASE_URL
	// (Viper sometimes misses it when no .env file exists)
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	}

	return &cfg, nil
}

// DSN returns the PostgreSQL connection string.
// Prefers DATABASE_URL (Railway/Render/Heroku style) over individual fields.
func (c *Config) DSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}
