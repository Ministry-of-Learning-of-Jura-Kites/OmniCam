package messages_workspace

import (
	"encoding/json"

	"github.com/google/uuid"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
)

type WorkspaceNoCams struct {
	ModelId     uuid.UUID `json:"modelId"`
	UserId      uuid.UUID `json:"userId"`
	Version     int32     `json:"version"`
	BaseVersion int32     `json:"baseVersion"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

type Workspace struct {
	WorkspaceNoCams
	Cameras     *messages_cameras.Cameras `json:"cameras"`
	BaseCameras *messages_cameras.Cameras `json:"baseCameras"`
}

type WorkspaceRawCams struct {
	WorkspaceNoCams
	Cameras     json.RawMessage `json:"cameras"`
	BaseCameras json.RawMessage `json:"baseCameras"`
}
