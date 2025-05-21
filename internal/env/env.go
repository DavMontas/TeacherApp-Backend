package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {

	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	// Atoi is equivalent to ParseInt(s, 10, 0), converted to type int
	result, err := strconv.Atoi(val)

	if err != nil {
		return fallback
	}

	return result
}
