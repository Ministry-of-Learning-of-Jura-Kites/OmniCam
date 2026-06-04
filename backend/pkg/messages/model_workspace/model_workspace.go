package messages_model_workspace

import (
	"github.com/google/uuid"
	messages_cameras "omnicam.com/backend/pkg/messages/cameras"
	messages_trapezoids "omnicam.com/backend/pkg/messages/trapezoids"
)

type Simulation struct {
	Areas            []SimulationArea  `json:"areas"`
	PopulationGroups []PopulationGroup `json:"populationGroups"`
}

type SimulationArea struct {
	Id     string       `json:"id"`
	Name   string       `json:"name"`
	Color  string       `json:"color"`
	Points [][3]float64 `json:"points"`
	Kind   string       `json:"kind"`
}

type PopulationGroup struct {
	Id string `json:"id"`

	StartAreaId string `json:"startAreaId"`
	EndAreaId   string `json:"endAreaId"`

	Height float64 `json:"height"`
	Speed  float64 `json:"speed"`
	Count  int32   `json:"count"`
}

type ModelWorkspace struct {
	ModelId          uuid.UUID                       `json:"modelId"`
	Name             string                          `json:"name"`
	Description      string                          `json:"description"`
	Version          int32                           `json:"version"`
	CreatedAt        string                          `json:"createdAt"`
	UpdatedAt        string                          `json:"updatedAt"`
	Cameras          *messages_cameras.Cameras       `json:"cameras"`
	TargetTrapezoids *messages_trapezoids.Trapezoids `json:"targetTrapezoids"`
	ScaleFactor      float64                         `json:"scaleFactor"`
	ModelHeight      float64                         `json:"modelHeight"`
	ProjectId        uuid.UUID                       `json:"projectId"`
	FilePath         string                          `json:"filePath"`
	ModelExtension   string                          `json:"fileExtension"`
	ImagePath        string                          `json:"imagePath"`
	ImageExtension   string                          `json:"imageExtension"`
	Simulation       Simulation                      `json:"simulation"`
}
