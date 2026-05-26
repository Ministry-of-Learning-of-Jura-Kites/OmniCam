package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWT(firstName string, lastName string, userID string, username string, jwtSecret string, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := UserClaims{
		UserID:    userID,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
