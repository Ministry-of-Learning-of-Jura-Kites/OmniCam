package messages_cameras

import camera "omnicam.com/backend/pkg/messages/protobufs"

type Cameras = map[string]CameraStruct

type ColorRGBA struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

type CameraStruct struct {
	Name              string    `json:"name" binding:"required" diff:"name"`
	AngleX            float64   `json:"angleX" binding:"required" diff:"angleX"`
	AngleY            float64   `json:"angleY" binding:"required" diff:"angleY"`
	AngleZ            float64   `json:"angleZ" binding:"required" diff:"angleZ"`
	AngleW            float64   `json:"angleW" binding:"required" diff:"angleW"`
	PosX              float64   `json:"posX" binding:"required" diff:"posX"`
	PosY              float64   `json:"posY" binding:"required" diff:"posY"`
	PosZ              float64   `json:"posZ" binding:"required" diff:"posZ"`
	Fov               float64   `json:"fov" binding:"required" diff:"fov"`
	FrustumColor      ColorRGBA `json:"frustumColor" binding:"required" diff:"frustumColor"`
	FrustumLength     float64   `json:"frustumLength" binding:"required" diff:"frustumLength"`
	IsHidingArrows    bool      `json:"isHidingArrows" diff:"-"`
	IsHidingWheels    bool      `json:"isHidingWheels" diff:"-"`
	IsLockingPosition bool      `json:"isLockingPosition" diff:"-"`
	IsLockingRotation bool      `json:"isLockingRotation" diff:"-"`
	IsHidingFrustum   bool      `json:"isHidingFrustum" diff:"-"`
	AspectWidth       float64   `json:"aspectWidth" binding:"required" diff:"aspectWidth"`
	AspectHeight      float64   `json:"aspectHeight" binding:"required" diff:"aspectHeight"`
}

func ProtoColorToColor(color *camera.ColorRGBA) ColorRGBA {
	return ColorRGBA{
		R: color.R,
		G: color.G,
		B: color.B,
		A: color.A,
	}
}

func ProtoCamToCam(protoCam *camera.Camera) CameraStruct {
	return CameraStruct{
		Name:              protoCam.Name,
		AngleX:            protoCam.AngleX,
		AngleY:            protoCam.AngleY,
		AngleZ:            protoCam.AngleZ,
		AngleW:            protoCam.AngleW,
		PosX:              protoCam.PosX,
		PosY:              protoCam.PosY,
		PosZ:              protoCam.PosZ,
		Fov:               protoCam.Fov,
		FrustumColor:      ProtoColorToColor(protoCam.FrustumColor),
		FrustumLength:     protoCam.FrustumLength,
		IsHidingArrows:    protoCam.IsHidingArrows,
		IsHidingWheels:    protoCam.IsHidingWheels,
		IsLockingPosition: protoCam.IsLockingPosition,
		IsLockingRotation: protoCam.IsLockingRotation,
		IsHidingFrustum:   protoCam.IsHidingFrustum,
		AspectWidth:       protoCam.AspectWidth,
		AspectHeight:      protoCam.AspectHeight,
	}
}
