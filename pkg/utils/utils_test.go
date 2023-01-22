package utils

import (
	"fmt"
	"reflect"
	"testing"
)

type TestStruct struct {

}

func SanitizeMessageValueStruct(t *testing.T) {
	testStruct := TestStruct{}
	v := SanitizeMessageValue(testStruct)
	ans := reflect.ValueOf(v).Kind()

	expected := reflect.String

	if ans != expected {
		t.Errorf("Failed message value sanitization. Expected %s. Got %s", expected, ans)
	}
}

func SanitizeMessageValueBool(t *testing.T) {
	testBool := false
	v := SanitizeMessageValue(testBool)
	ans := reflect.ValueOf(v).Kind()

	expected := reflect.Bool

	if ans != expected {
		t.Errorf("Failed message value sanitization. Expected %s. Got %s", expected, ans)
	}
}

func PositiveSanitizeLogMessageTest(t *testing.T) {
	userInput := "5ca9dba"
	log := fmt.Sprintf("This is a log line with a cache key: %s", userInput)
	expected := "This is a log line with a cache key: 5ca****"

	ans := SanitizeLogMessage(log, userInput)

	if ans != expected {
		t.Errorf("The sanitize log message does not match. The response was: %s", ans)
	}
}

func PositiveSanitizeLogMessageMultipleTest(t *testing.T) {
	userInput := "5ca9dba"
	log := fmt.Sprintf("This is a log line with a cache key: %s", userInput)
	expected := "This is a log line with a cache key: 5ca****. Cache key: 5ca****"

	ans := SanitizeLogMessage(log, userInput)

	if ans != expected {
		t.Errorf("The sanitize log message does not match. The response was: %s", ans)
	}
}

func PositiveSanitizeLogMessageShortKeyTest(t *testing.T) {
	userInput := "5c"
	log := fmt.Sprintf("This is a log line with a cache key: %s", userInput)
	expected := fmt.Sprintf("This is a log line with a cache key: %s****", userInput)

	ans := SanitizeLogMessage(log, userInput)

	if ans != expected {
		t.Errorf("The sanitize log message does not match. The response was: %s", ans)
	}
}

func PositiveSanitizeLogMessageEmptyTest(t *testing.T) {
	userInput := ""
	log := fmt.Sprintf("This is a log line with a cache key: %s", userInput)
	expected := fmt.Sprintf("This is a log line with a cache key: %s****", userInput)

	ans := SanitizeLogMessage(log, userInput)

	if ans != expected {
		t.Errorf("The sanitize log message does not match. The response was: %s", ans)
	}
}