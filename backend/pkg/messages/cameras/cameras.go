package messages_cameras

import (
	"encoding/json"

	camera "omnicam.com/backend/pkg/messages/protobufs"
)

type CamId string

type Cameras = map[CamId]CameraStruct

type ColorRGBA struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

type DistortionConfig struct {
	Enabled   bool `json:"enabled"`
	IsFisheye bool `json:"isFisheye"`
}

type CameraStruct struct {
	Name              string           `json:"name" binding:"required" diff:"name"`
	AngleX            float64          `json:"angleX" diff:"angleX"`
	AngleY            float64          `json:"angleY" diff:"angleY"`
	AngleZ            float64          `json:"angleZ" diff:"angleZ"`
	AngleW            float64          `json:"angleW" diff:"angleW"`
	PosX              float64          `json:"posX" diff:"posX"`
	PosY              float64          `json:"posY" diff:"posY"`
	PosZ              float64          `json:"posZ" diff:"posZ"`
	Fov               float64          `json:"fov"  diff:"fov"`
	FrustumColor      ColorRGBA        `json:"frustumColor" diff:"frustumColor"`
	FrustumLength     float64          `json:"frustumLength" diff:"frustumLength"`
	IsHidingArrows    bool             `json:"isHidingArrows" diff:"-"`
	IsHidingWheels    bool             `json:"isHidingWheels" diff:"-"`
	IsLockingPosition bool             `json:"isLockingPosition" diff:"-"`
	IsLockingRotation bool             `json:"isLockingRotation" diff:"-"`
	IsHidingFrustum   bool             `json:"isHidingFrustum" diff:"-"`
	AspectWidth       float64          `json:"aspectWidth" binding:"required" diff:"aspectWidth"`
	AspectHeight      float64          `json:"aspectHeight" binding:"required" diff:"aspectHeight"`
	Distortion        DistortionConfig `json:"distortion" diff:"distortion"`
}

func UnmarshalCameras(data []byte) (Cameras, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	result := make(Cameras)
	for id, cameraBlob := range raw {
		// Start with a fresh set of default values for EACH camera
		cam := DefaultCam()

		// Unmarshal the JSON blob into the defaulted struct.
		// json.Unmarshal only overwrites fields that ARE present in the JSON.
		if err := json.Unmarshal(cameraBlob, &cam); err != nil {
			return nil, err
		}

		result[CamId(id)] = cam
	}

	return result, nil
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
		Enabled:   true,
		IsFisheye: false,
	}
	if protoDist != nil {
		distortion = DistortionConfig{
			Enabled:   protoDist.Enabled,
			IsFisheye: protoDist.IsFisheye,
		}
	}
	return distortion
}

func ProtoCamToCam(protoCam *camera.Camera) CameraStruct {
	cam := DefaultCam()

	cam.Name = protoCam.Name
	cam.AngleX = protoCam.AngleX
	cam.AngleY = protoCam.AngleY
	cam.AngleZ = protoCam.AngleZ
	cam.AngleW = protoCam.AngleW
	cam.PosX = protoCam.PosX
	cam.PosY = protoCam.PosY
	cam.PosZ = protoCam.PosZ
	cam.Fov = protoCam.Fov
	cam.FrustumLength = protoCam.FrustumLength
	cam.IsHidingArrows = protoCam.IsHidingArrows
	cam.IsHidingWheels = protoCam.IsHidingWheels
	cam.IsLockingPosition = protoCam.IsLockingPosition
	cam.IsLockingRotation = protoCam.IsLockingRotation
	cam.IsHidingFrustum = protoCam.IsHidingFrustum
	cam.AspectWidth = protoCam.AspectWidth
	cam.AspectHeight = protoCam.AspectHeight

	if protoCam.FrustumColor != nil {
		cam.FrustumColor = ColorRGBA{
			R: protoCam.FrustumColor.R,
			G: protoCam.FrustumColor.G,
			B: protoCam.FrustumColor.B,
			A: protoCam.FrustumColor.A,
		}
	}

	if protoCam.Distortion != nil {
		cam.Distortion = DistortionConfig{
			Enabled:   protoCam.Distortion.Enabled,
			IsFisheye: protoCam.Distortion.IsFisheye,
		}
	}

	return cam
}

func DefaultCam() CameraStruct {
	return CameraStruct{
		Name:              "Untitled",
		AngleX:            0,
		AngleY:            0,
		AngleZ:            0,
		AngleW:            0,
		PosX:              0,
		PosY:              0,
		PosZ:              0,
		Fov:               60,
		FrustumColor:      ColorRGBA{R: 0.5, G: 0.5, B: 0.5, A: 0.5},
		FrustumLength:     1000,
		IsHidingArrows:    false,
		IsHidingWheels:    false,
		IsLockingPosition: false,
		IsLockingRotation: false,
		IsHidingFrustum:   true,
		AspectWidth:       1,
		AspectHeight:      1,
		Distortion: DistortionConfig{
			Enabled:   true,
			IsFisheye: false,
		},
	}
}
