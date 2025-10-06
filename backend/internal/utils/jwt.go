package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(first_name string, last_name string, userID string, username string, jwtSecret string, jwtExpireTime int32) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"first_name": first_name,
		"last_name": last_name,
		"username": username,
		"exp":      time.Now().Add(time.Duration(jwtExpireTime)).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
