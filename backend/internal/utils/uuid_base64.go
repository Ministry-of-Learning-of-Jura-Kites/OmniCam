package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseUuidBase64(strId string) (uuid.UUID, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(strId)
	if err != nil {
		return uuid.Nil, err
	}
	parsedId, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		return uuid.Nil, err
	}
	return parsedId, nil
}

func GetUuidFromCtx(c *gin.Context, key string) (uuid.UUID, error) {
	anyUserId, exists := c.Get(key)
	if !exists {
		return uuid.Nil, fmt.Errorf("cannot find %s from context", key)
	}
	userId, ok := anyUserId.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("cannot cast uuid value of key %s to uuid.UUID", key)
	}

	return userId, nil
}

func UuidToBase64(id uuid.UUID) (string, error) {
	strId, err := id.MarshalBinary()
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(strId), nil
}
