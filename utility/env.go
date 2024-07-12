package utility

import "os"

func GetEnvWithDefault(envStr, defaultValue string) string {
	if env := os.Getenv(envStr); env != "" {
		return env
	}
	return defaultValue
}
