package utils

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidations() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("utf8only", func(fl validator.FieldLevel) bool {
			s := strings.TrimSpace(fl.Field().String())
			return s != "" && utf8.ValidString(s)
		})
	}
}

func IsValidUsername(username string) bool {
	for _, ch := range username {
		if !isEnglishAlphabet(ch) && !unicode.IsNumber(ch) && !slices.Contains([]rune{'-', '_', '.'}, ch) {
			return false
		}
	}
	return true
}
