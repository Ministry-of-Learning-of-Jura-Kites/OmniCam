package messages_cameras

type Cameras = map[string]CameraStruct

type CameraStruct struct {
	Name           string  `json:"name" binding:"required"`
	AngleX         float64 `json:"angleX" binding:"required"`
	AngleY         float64 `json:"angleY" binding:"required"`
	AngleZ         float64 `json:"angleZ" binding:"required"`
	AngleW         float64 `json:"angleW" binding:"required"`
	PosX           float64 `json:"posX" binding:"required"`
	PosY           float64 `json:"posY" binding:"required"`
	PosZ           float64 `json:"posZ" binding:"required"`
	Fov            float64 `json:"fov" binding:"required"`
	IsHidingArrows bool    `json:"isHidingArrows"`
	IsHidingWheels bool    `json:"isHidingWheels"`
}
