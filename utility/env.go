package utility

import "os"

func GetEnvWithDefault(envStr, defaultValue string) string {
	if env := os.Getenv(envStr); env != "" {
		return env
	}
	return defaultValue
}

// IsDevChan checks if the current environment is a development channel
func IsDevChan() bool {
	return os.Getenv("DEV_CHAN") != ""
}
