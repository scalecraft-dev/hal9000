package internal

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("ENVIRONMENT") != "production" {
			fmt.Println("Warning: Error loading .env file:", err)
		}
	}

	config := &Config{
		SlackToken:         getEnv("SLACK_TOKEN", "", true),
		SlackSigningSecret: getEnv("SLACK_SIGNING_SECRET", "", true),
		ServerPort:         getEnvAsInt("SERVER_PORT", 50051, false),
		ServerHost:         getEnv("SERVER_HOST", "0.0.0.0", false),
		Environment:        getEnv("ENVIRONMENT", "development", false),
		LogLevel:           getEnv("LOG_LEVEL", "info", false),
	}

	return config, nil
}

func getEnv(key, defaultValue string, required bool) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		if required {
			panic(fmt.Sprintf("Required environment variable %s is not set", key))
		}
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int, required bool) int {
	valueStr := getEnv(key, "", required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		if required {
			panic(fmt.Sprintf("Environment variable %s must be an integer", key))
		}
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool, required bool) bool {
	valueStr := getEnv(key, "", required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		if required {
			panic(fmt.Sprintf("Environment variable %s must be a boolean", key))
		}
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration, required bool) time.Duration {
	valueStr := getEnv(key, "", required)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		if required {
			panic(fmt.Sprintf("Environment variable %s must be a valid duration", key))
		}
		return defaultValue
	}
	return value
}
