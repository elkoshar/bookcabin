package helpers

import (
	"os"
)

// Environment List
const (
	EnvLocal       = "local"
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// GetEnv abstract on top of standard `os` to have fallback value
func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

// Get return string of current environment flag
func GetEnvString() string {
	env := os.Getenv("GO_ENV")

	if env == "" {
		env = EnvLocal
	}

	return env
}
