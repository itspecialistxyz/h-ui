package middleware

import (
	"os"

	"h-ui/service"

	"github.com/sirupsen/logrus"
)

// Helper to get config value from DB or env
func GetConfigValue(key string) (string, error) {
	if key == "" {
		logrus.Error("GetConfigValue called with empty key")
		return "", nil
	}

	// First try to get value from database
	config, err := service.GetConfig(key)
	if err == nil && config.Value != nil && *config.Value != "" {
		return *config.Value, nil
	} else if err != nil {
		logrus.Debugf("Error getting config from DB for key %s: %v, falling back to environment variable", key, err)
	}

	// Fall back to environment variable
	val := os.Getenv(key)
	return val, nil
}
