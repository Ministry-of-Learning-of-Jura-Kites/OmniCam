package utils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UuidToPgUuid(uuid uuid.UUID) (pgtype.UUID, error) {
	uuidBytes, err := uuid.MarshalBinary()
	if err != nil {
		return pgtype.UUID{Valid: false}, fmt.Errorf("error while marshaling uuid: %s", err)
	}

	return pgtype.UUID{
		Bytes: [16]byte(uuidBytes),
		Valid: true,
	}, nil
}
