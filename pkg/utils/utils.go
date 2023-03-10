// Package utils for utility functions
package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// SanitizeMongoInput sanitize Mongo input to guard against
// NoSQL injection
func SanitizeMongoInput(s string) string {
	m1 := regexp.MustCompile(`/^\$|\./g`)
	// return strings.Trim(s, " $/^\\")
	return m1.ReplaceAllString(s, "-")
}

// MapToProtobufStruct convert a map to a struct. This helps to
// encoding the struct to JSON when a message is consumed.
func MapToProtobufStruct(m map[string]interface{}) (*structpb.Struct, error) {
	s, err := structpb.NewStruct(m)
	return s, err
}

// SanitizeMessageValue ensures that a message's value is an
// int, int64, bool, string
func SanitizeMessageValue(i interface{}) interface{} {
	switch v := i.(type) {
	case int:
	case string:
	case int64:
	case bool:
		return v
	default:
		sv := fmt.Sprintf("%v", v)
		return sv
	}

	return nil
}

// SanitizeLogMessage removes user input from the log output
func SanitizeLogMessage(log string, input ...string) string {
	cleanLog := log

	for _, v := range input {
		if v == "" {
			return log
		}

		cleanUserInput := "****"
		if len(v) > 3 {
			cleanUserInput = v[0:3] + "****"
		}

		//fullLog := fmt.Sprintf(log, userInput)
		cleanLog = strings.Replace(cleanLog, v, cleanUserInput, -1)
	}
	cleanLog = strings.Replace(cleanLog, "\n", "", -1)
	cleanLog = strings.Replace(cleanLog, "\r", "", -1)

	return cleanLog
}

// SanitizeLogMessageF removes user input from the log output and format
func SanitizeLogMessageF(log string, input ...string) string {
	cleanLog := log

	for _, v := range input {
		cleanLog := fmt.Sprintf(log, v)

		if v == "" {
			return log
		}

		cleanUserInput := "****"
		if len(v) > 3 {
			cleanUserInput = v[0:3] + "****"
		}

		//fullLog := fmt.Sprintf(log, userInput)
		cleanLog = strings.Replace(cleanLog, v, cleanUserInput, -1)
	}
	cleanLog = strings.Replace(cleanLog, "\n", "", -1)
	cleanLog = strings.Replace(cleanLog, "\r", "", -1)

	return cleanLog
}

// SanitizeErrorMessage removes user input from the err output
func SanitizeErrorMessage(log error, input ...string) error {
	cleanLog := log.Error()

	for _, v := range input {
		if v == "" {
			return log
		}

		cleanUserInput := "****"
		if len(v) > 3 {
			cleanUserInput = v[0:3] + "****"
		}

		//fullLog := fmt.Sprintf(log, userInput)
		cleanLog = strings.Replace(cleanLog, v, cleanUserInput, -1)
	}
	cleanLog = strings.Replace(cleanLog, "\n", "", -1)
	cleanLog = strings.Replace(cleanLog, "\r", "", -1)

	return fmt.Errorf("%s", cleanLog)
}

// ObfuscateValue obfuscate the string
func ObfuscateValue(input string, char int) string {
	if input == "" || char < 0 {
		return input
	}

	cleanUserInput := "****"
	if len(input) > 3 {
		cleanUserInput = input[0:char] + "****"
	}

	return cleanUserInput
}

// GetEnv get key environment variable if exist otherwise return defalutValue
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
