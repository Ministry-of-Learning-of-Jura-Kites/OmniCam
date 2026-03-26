package target_trapezoids

import (
	"omnicam.com/backend/pkg/messages/protobufs"
)

type TrapezoidId string

type Trapezoids = map[TrapezoidId]TrapezoidStruct

type TrapezoidStruct struct {
	ID     string     `json:"id" binding:"required" diff:"id"`
	Name   string     `json:"name" diff:"name"`
	Points [4]Vector3 `json:"points" diff:"points"`
	Color  *string    `json:"color,omitempty" diff:"color"`
	Hidden bool       `json:"hidden" diff:"hidden"`
}

// Vector3 mirrors ProtoVector3
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func ProtoTrapezoidToTrapezoid(protoCam *protobufs.CoverageFace) TrapezoidStruct {
	if protoCam == nil {
		return TrapezoidStruct{}
	}

	// Map points slice
	var points [4]Vector3
	for i, p := range protoCam.Points {
		points[i] = Vector3{X: p.X, Y: p.Y, Z: p.Z}
	}

	return TrapezoidStruct{
		ID:     protoCam.Id,
		Name:   protoCam.Name,
		Points: points,
		Color:  protoCam.Color, // proto3 optional fields are generated as *string in Go
		Hidden: protoCam.Hidden,
	}
}
