package utility

func GetEnvWithDefault(envStr, defaultValue string) string {
	if envStr == "" {
		return defaultValue
	}
	return envStr
}
