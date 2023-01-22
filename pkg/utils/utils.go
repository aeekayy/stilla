// Package utils for utility functions
package utils

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// SanitizeMongoInput sanitize Mongo input to guard against
// NoSQL injection
func SanitizeMongoInput(s string) string {
	return strings.Trim(s, " $/^\\")
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
func SanitizeLogMessage(log, userInput string) string {
	cleanUserInput := userInput + "****"
	if len(userInput) > 3 {
		cleanUserInput = userInput[0:3] + "****"
	}
	return strings.Replace(log, userInput, cleanUserInput, -1)
}