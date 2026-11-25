package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	JWTTTLMinutes       int
	RefreshTokenTTLDays int
	S3Endpoint          string
	S3AccessKey         string
	S3SecretKey         string
	S3Bucket            string
	S3PublicBaseURL     string
	S3UsePathStyle      bool
}

func Load() *Config {
	jwtTTL, _ := strconv.Atoi(getEnv("JWT_TTL_MINUTES", "15"))
	refreshTTL, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_TTL_DAYS", "7"))
	pathStyle, _ := strconv.ParseBool(getEnv("S3_USE_PATH_STYLE", "true"))

	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://imovxen_user:imovxen_pass@localhost:5432/imovxen?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "super-secret-key-change-in-production-min32chars"),
		JWTTTLMinutes:       jwtTTL,
		RefreshTokenTTLDays: refreshTTL,
		S3Endpoint:          getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3AccessKey:         getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:         getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:            getEnv("S3_BUCKET", "imovxen-images"),
		S3PublicBaseURL:     getEnv("S3_PUBLIC_BASE_URL", "http://localhost:9000/imovxen-images"),
		S3UsePathStyle:      pathStyle,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
