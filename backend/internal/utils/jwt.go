package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(name string, surname string, userID string, username string, jwtSecret string, jwtExpireTime int32) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"name":     name,
		"surname":  surname,
		"username": username,
		"exp":      time.Now().Add(time.Duration(jwtExpireTime)).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
