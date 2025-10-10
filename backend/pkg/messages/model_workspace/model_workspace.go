package messages_model_workspace

import (
	"github.com/google/uuid"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
)

type ModelWorkspace struct {
	ModelId     uuid.UUID                 `json:"modelId"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Version     int32                     `json:"version"`
	CreatedAt   string                    `json:"createdAt"`
	UpdatedAt   string                    `json:"updatedAt"`
	Cameras     *messages_cameras.Cameras `json:"cameras"`
	ProjectId   uuid.UUID                 `json:"projectId"`
	FilePath    string                    `json:"filePath"`
	ImagePath   string                    `json:"imagePath"`
}
