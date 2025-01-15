package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	Port        string
	Environment string
	APIVersion  string

	// JWT Configuration
	JWTSecret     string
	JWTExpiration time.Duration

	// Database Configuration
	MongoURI  string
	MongoDB   string
	MongoUser string
	MongoPass string

	// Redis Configuration
	RedisURI      string
	RedisPassword string

	// External Services
	ITunesAPIURL      string
	ChartLyricsAPIURL string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Si .env no existe, intentamos cargar desde las variables de ambiente
		// No retornamos error porque en producción usaremos variables de ambiente
	}

	jwtExpStr := getEnv("JWT_EXPIRATION", "24h")
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENV", "development"),
		APIVersion:        getEnv("API_VERSION", "v1"),
		JWTSecret:         getEnv("JWT_SECRET", "your_jwt_secret_here"),
		JWTExpiration:     jwtExp,
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:           getEnv("MONGO_DB_NAME", "music_api"),
		MongoUser:         getEnv("MONGO_USER", ""),
		MongoPass:         getEnv("MONGO_PASSWORD", ""),
		RedisURI:          getEnv("REDIS_URI", "redis://localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		ITunesAPIURL:      getEnv("ITUNES_API_URL", "https://itunes.apple.com"),
		ChartLyricsAPIURL: getEnv("CHARTLYRICS_API_URL", "http://api.chartlyrics.com/apiv1.asmx"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
