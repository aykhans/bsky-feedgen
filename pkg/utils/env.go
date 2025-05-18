package utils

import (
	"fmt"
	"os"
)

// GetEnv retrieves the environment variable named by the key envName,
// and attempts to parse its value into the specified type T.
//
// It returns the parsed value of type T and a nil error on success.
// It returns a zero value of type T and an error if the environment
// variable is not set, or if parsing fails using the ParseValue function.
func GetEnv[T any](envName string) (T, error) {
	var zero T

	envStr := os.Getenv(envName)
	if envStr == "" {
		return zero, fmt.Errorf("environment variable %s is not set", envName)
	}

	parsedEnv, err := ParseString[T](envStr)
	if err != nil {
		return zero, fmt.Errorf("failed to parse environment variable %s: %s", envName, err)
	}

	return parsedEnv, nil
}

// GetEnvOr retrieves the environment variable named by the key envName,
// and attempts to parse its value into the specified type T.
//
// It returns the parsed value of type T and a nil error on success.
// If the environment variable is not set, it returns the provided default value
// and a nil error. If parsing fails, it returns a zero value of type T and an error.
func GetEnvOr[T any](envName string, defaultValue T) (T, error) {
	var zero T

	envStr := os.Getenv(envName)
	if envStr == "" {
		return defaultValue, nil
	}

	parsedEnv, err := ParseString[T](envStr)
	if err != nil {
		return zero, fmt.Errorf("failed to parse environment variable %s: %s", envName, err)
	}

	return parsedEnv, nil
}
