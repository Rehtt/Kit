package util

import "os"

func Getenv(key string, defaultValue string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	return defaultValue
}
