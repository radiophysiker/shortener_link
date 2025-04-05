package utils

import (
	"testing"
	"unicode/utf8"
)

func TestGetShortRandomString(t *testing.T) {
	length := 10
	result := GetShortRandomString(length)

	if utf8.RuneCountInString(result) != length {
		t.Errorf("Expected string of length %d, but got %d", length, utf8.RuneCountInString(result))
	}

	// Check if the string is unique by running the function again
	secondResult := GetShortRandomString(length)
	if result == secondResult {
		t.Errorf("Expected unique strings, but got two identical ones")
	}
}
