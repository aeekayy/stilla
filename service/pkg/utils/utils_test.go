package utils

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
}

func TestSanitizeMessageValueStruct(t *testing.T) {
	testStruct := TestStruct{}
	v := SanitizeMessageValue(testStruct)
	ans := reflect.ValueOf(v).Kind()

	expected := reflect.String

	if ans != expected {
		t.Errorf("Failed message value sanitization. Expected %s. Got %s", expected, ans)
	}
}

// TestSuiteSanitizeMessage
func TestSuiteSanitizeMessage(t *testing.T) {
	userInput := "5ca9dba"
	shortUserInput := "5c"
	emptyUserInput := ""

	table := []struct {
		name     string
		input    string
		log      string
		expected string
	}{
		{"TestPositiveSanitizeLogMessage", userInput, fmt.Sprintf("This is a log line with a cache key: %s", userInput), "This is a log line with a cache key: 5ca****"},
		{"TestPositiveSanitizeLogMessageMultiple", userInput, fmt.Sprintf("This is a log line with a cache key: %s. Cache key: %s", userInput, userInput), "This is a log line with a cache key: 5ca****. Cache key: 5ca****"},
		{"TestPositiveSanitizeLogMessageShortKey", shortUserInput, fmt.Sprintf("This is a log line with a cache key: %s", shortUserInput), "This is a log line with a cache key: ****"},
		{"TestPositiveSanitizeLogMessageEmpty", emptyUserInput, fmt.Sprintf("This is a log line with a cache key: %s", emptyUserInput), "This is a log line with a cache key: "},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			ans := SanitizeLogMessage(tc.log, tc.input)
			assert.Equal(t, ans, tc.expected, "the log outputs should match.")
		})
	}
}

// TestSanitizeMessageValueBool
func TestSanitizeMessageValueBool(t *testing.T) {
	testBool := false
	v := SanitizeMessageValue(testBool)
	ans := reflect.ValueOf(v).Kind()

	expected := reflect.Bool

	if ans != expected {
		t.Errorf("Failed message value sanitization. Expected %s. Got %s", expected, ans)
	}
}

// TestSuiteGetEnv
func TestSuiteGetEnv(t *testing.T) {
	table := []struct {
		name          string
		inputKey      string
		inputDefault  string
		inputValue    string
		expectedValue string
	}{
		{"TestPositiveGetEnv", "Test", "test1", "test2", "test2"},
		{"TestEmptyGetEnv", "", "localhost", "", "localhost"},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			if tc.inputKey != "" {
				err := os.Setenv(tc.inputKey, tc.inputValue)

				if err != nil {
					t.Errorf("error setting environment variable: %s", err)
				}

				ans := GetEnv(tc.inputKey, tc.inputDefault)
				assert.Equal(t, tc.expectedValue, ans, "the environment variable values should match.")
			}
		})
	}
}
