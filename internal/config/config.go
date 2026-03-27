package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config is a single place that holds all environment-driven settings.
//
// Why we do this:
// - main.go wires dependencies (DB, JWT, MinIO) using values from Config
// - handlers/services should not read env directly (keeps code testable)
type Config struct {
	Port string

	DBURL string

	JWTSecret string
	JWTExpiry time.Duration

	// PublicBaseURL is used to build public links for uploaded files.
	//
	// Examples:
	// - local:   http://localhost:8080
	// - railway: https://city-service-production.up.railway.app
	PublicBaseURL string

	// StorageDriver selects where we store uploaded photos.
	// - "minio": upload to S3-compatible storage (MinIO)
	// - "local": save to ./uploads and serve via /uploads/*
	StorageDriver string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

// Load reads .env (if present) + environment variables and returns Config.
//
// Note:
//   - In production, .env file may not exist; env vars are usually injected by the platform.
//   - We still call godotenv.Load() because it's convenient locally; it does NOT error if file is missing
//     (it only errors if it exists but cannot be read).
func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:          getenv("PORT", "8080"),
		DBURL:         os.Getenv("DB_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		PublicBaseURL: os.Getenv("PUBLIC_BASE_URL"),
		StorageDriver: getenv("STORAGE_DRIVER", ""),

		MinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    getenv("MINIO_BUCKET", "city-service"),
	}

	if cfg.DBURL == "" {
		return Config{}, fmt.Errorf("DB_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	// JWT_EXPIRY is a duration string like "24h" or "15m".
	expiryStr := getenv("JWT_EXPIRY", "24h")
	exp, err := time.ParseDuration(expiryStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_EXPIRY %q: %w", expiryStr, err)
	}
	cfg.JWTExpiry = exp

	// MINIO_USE_SSL is "true"/"false" in the .env.
	useSSLStr := getenv("MINIO_USE_SSL", "false")
	useSSL, err := strconv.ParseBool(useSSLStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid MINIO_USE_SSL %q: %w", useSSLStr, err)
	}
	cfg.MinioUseSSL = useSSL

	// Storage driver defaulting:
	// - If STORAGE_DRIVER is set, we trust it.
	// - Otherwise: if MINIO_ENDPOINT exists, use MinIO; else use local storage.
	if cfg.StorageDriver == "" {
		if cfg.MinioEndpoint != "" {
			cfg.StorageDriver = "minio"
		} else {
			cfg.StorageDriver = "local"
		}
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
