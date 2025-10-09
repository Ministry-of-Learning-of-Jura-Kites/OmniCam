package utils

import (
	"encoding/base64"

	"github.com/google/uuid"
)

func ParseUuidBase64(strId string) (*uuid.UUID, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(strId)
	if err != nil {
		return nil, err
	}
	parsedId, err := uuid.FromBytes(decodedBytes)
	if err != nil {
		return nil, err
	}
	return &parsedId, nil
}
