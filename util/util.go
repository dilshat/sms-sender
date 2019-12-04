package util

import (
	"os"
	"strconv"
	"strings"
	"unicode"
)

func FileExists(name string) bool {
	_, err := os.Stat(name)

	if os.IsNotExist(err) {
		return false
	}

	//sometimes there can be permission or other errors
	//here we use a simple logic that if file exists and we can use it then true otherwise false
	return err == nil
}

func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func GetEnvAsInt(name string, defaultVal int) int {
	valueStr := GetEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
