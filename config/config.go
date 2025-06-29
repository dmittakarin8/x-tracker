package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"path/filepath"

	"github.com/joho/godotenv"
	"x-tracker/internal/logger"
)

type Config struct {
	// API Configuration
	RapidAPIKey      string
	RapidAPIHost     string
	RapidAPIEndpoint string
	
	// Rate Limiting
	MaxRequestsPerMinute int
	RequestTimeout       time.Duration
	
	// Database
	DBPath string
	
	// Discord Webhook (optional)
	DiscordWebhookURL string
	
	// Application Settings
	CheckInterval time.Duration
	
	// Logging
	LoggingEnabled bool
	LogDir         string

	// Notification Controls
	EnableFollowNotifications   bool
	EnableUnfollowNotifications bool
	EnableDiscordNotifications  bool
	EnableTelegramNotifications bool

	// Webhook Configuration
	TelegramBotToken string
	TelegramChatID   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Default database path in user's home directory
	defaultDBPath := filepath.Join(homeDir, ".x-tracker", "data.db")

	maxRequests, _ := strconv.Atoi(getEnvWithDefault("MAX_REQUESTS_PER_MINUTE", "30"))
	checkIntervalStr := getEnvWithDefault("CHECK_INTERVAL", "5m")
	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid check interval format: %w", err)
	}
	logger.Info("Loaded check interval: %s", checkInterval)
	requestTimeout, _ := time.ParseDuration(getEnvWithDefault("REQUEST_TIMEOUT", "10s"))

	loggingEnabled, _ := strconv.ParseBool(getEnvWithDefault("LOGGING_ENABLED", "false"))

	return &Config{
		RapidAPIKey:         os.Getenv("RAPID_API_KEY"),
		RapidAPIHost:        os.Getenv("RAPID_API_HOST"),
		MaxRequestsPerMinute: maxRequests,
		RequestTimeout:       requestTimeout,
		DBPath:              getEnvWithDefault("DB_PATH", defaultDBPath),
		DiscordWebhookURL:   os.Getenv("DISCORD_WEBHOOK_URL"),
		CheckInterval:       checkInterval,
		LoggingEnabled:      loggingEnabled,
		LogDir:              getEnvWithDefault("LOG_DIR", filepath.Join(homeDir, ".x-tracker", "logs")),
		EnableFollowNotifications:   getEnvBool("ENABLE_FOLLOW_NOTIFICATIONS", true),
		EnableUnfollowNotifications: getEnvBool("ENABLE_UNFOLLOW_NOTIFICATIONS", true),
		EnableDiscordNotifications:   getEnvBool("ENABLE_DISCORD_NOTIFICATIONS", true),
		EnableTelegramNotifications:  getEnvBool("ENABLE_TELEGRAM_NOTIFICATIONS", true),
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:      os.Getenv("TELEGRAM_CHAT_ID"),
	}, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	val = strings.ToLower(val)
	return val == "true" || val == "1" || val == "yes"
} 