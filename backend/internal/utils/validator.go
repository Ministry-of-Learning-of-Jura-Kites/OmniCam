package utils

import (
	"encoding/base64"
	"strings"
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

		_ = v.RegisterValidation("base64", func(fl validator.FieldLevel) bool {
			s := fl.Field().String()
			if s == "" {
				return false
			}
			_, err := base64.StdEncoding.DecodeString(s)
			return err == nil
		})
	}
}
