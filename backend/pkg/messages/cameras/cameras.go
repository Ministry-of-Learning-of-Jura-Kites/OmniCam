package messages_cameras

import camera "omnicam.com/backend/pkg/messages/protobufs"

type CamId string

type Cameras = map[CamId]CameraStruct

type ColorRGBA struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

type DistortionConfig struct {
	Intensity float64               `json:"intensity"`
	Mode      camera.DistortionMode `json:"mode"`
}

type CameraStruct struct {
	Name              string           `json:"name" binding:"required" diff:"name"`
	AngleX            float64          `json:"angleX" binding:"required" diff:"angleX"`
	AngleY            float64          `json:"angleY" binding:"required" diff:"angleY"`
	AngleZ            float64          `json:"angleZ" binding:"required" diff:"angleZ"`
	AngleW            float64          `json:"angleW" binding:"required" diff:"angleW"`
	PosX              float64          `json:"posX" binding:"required" diff:"posX"`
	PosY              float64          `json:"posY" binding:"required" diff:"posY"`
	PosZ              float64          `json:"posZ" binding:"required" diff:"posZ"`
	Fov               float64          `json:"fov" binding:"required" diff:"fov"`
	FrustumColor      ColorRGBA        `json:"frustumColor" binding:"required" diff:"frustumColor"`
	FrustumLength     float64          `json:"frustumLength" binding:"required" diff:"frustumLength"`
	IsHidingArrows    bool             `json:"isHidingArrows" diff:"-"`
	IsHidingWheels    bool             `json:"isHidingWheels" diff:"-"`
	IsLockingPosition bool             `json:"isLockingPosition" diff:"-"`
	IsLockingRotation bool             `json:"isLockingRotation" diff:"-"`
	IsHidingFrustum   bool             `json:"isHidingFrustum" diff:"-"`
	AspectWidth       float64          `json:"aspectWidth" binding:"required" diff:"aspectWidth"`
	AspectHeight      float64          `json:"aspectHeight" binding:"required" diff:"aspectHeight"`
	Distortion        DistortionConfig `json:"distortion" diff:"distortion"`
}

func ProtoColorToColor(protoColor *camera.ColorRGBA) ColorRGBA {
	color := ColorRGBA{R: 0.5, G: 0.5, B: 0.5, A: 0.5}
	if protoColor != nil {
		color = ColorRGBA{
			R: protoColor.R,
			G: protoColor.G,
			B: protoColor.B,
			A: protoColor.A,
		}
	}
	return color
}

func ProtoDistortionToDistortion(protoDist *camera.Distortion) DistortionConfig {
	distortion := DistortionConfig{
		Intensity: 1,
		Mode:      camera.DistortionMode_NONE,
	}
	if protoDist != nil {
		distortion = DistortionConfig{
			Intensity: protoDist.Intensity,
			Mode:      protoDist.Mode,
		}
	}
	return distortion
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
		Distortion:        ProtoDistortionToDistortion(protoCam.Distortion),
	}
}
