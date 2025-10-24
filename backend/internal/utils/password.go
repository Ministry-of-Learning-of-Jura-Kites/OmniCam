package utils

import (
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func CheckPasswordFormat(password string) bool {
	if len(password) < 8 || len(password) > 255 {
		return false
	}

	hasNumber := false
	hasSymbol := false

	for _, letter := range password {
		if unicode.IsNumber(letter) {
			hasNumber = true
		}
		if unicode.IsPunct(letter) || unicode.IsSymbol(letter) {
			hasSymbol = true
		}
	}

	return hasNumber && hasSymbol
}
