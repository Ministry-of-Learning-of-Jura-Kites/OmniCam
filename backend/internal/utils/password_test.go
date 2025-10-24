package utils_test

import (
	"testing"

	"omnicam.com/backend/internal/utils"
)

func TestCheckPasswordFormatPass(t *testing.T) {
	password := "Abcdef1!"
	if !utils.CheckPasswordFormat(password) {
		t.Errorf("Expected password to pass, but it failed")
	}
}

func TestCheckPasswordFormatFail(t *testing.T) {
	passwords := []string{
		"short1!",     // too short
		"NoNumber!",   // missing number
		"NoSymbol123", // missing symbol
		"",            // empty password
	}

	for _, pw := range passwords {
		t.Run(pw, func(t *testing.T) {
			if utils.CheckPasswordFormat(pw) {
				t.Errorf("Expected password %q to fail, but it passed", pw)
			}
		})
	}
}
