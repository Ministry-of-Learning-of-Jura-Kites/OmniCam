package messages_cameras

type Cameras = map[string]CameraStruct

type CameraStruct struct {
	Name              string  `json:"name" binding:"required" diff:"name"`
	AngleX            float64 `json:"angleX" binding:"required" diff:"angleX"`
	AngleY            float64 `json:"angleY" binding:"required" diff:"angleY"`
	AngleZ            float64 `json:"angleZ" binding:"required" diff:"angleZ"`
	AngleW            float64 `json:"angleW" binding:"required" diff:"angleW"`
	PosX              float64 `json:"posX" binding:"required" diff:"posX"`
	PosY              float64 `json:"posY" binding:"required" diff:"posY"`
	PosZ              float64 `json:"posZ" binding:"required" diff:"posZ"`
	Fov               float64 `json:"fov" binding:"required" diff:"fov"`
	IsHidingArrows    bool    `json:"isHidingArrows" diff:"-"`
	IsHidingWheels    bool    `json:"isHidingWheels" diff:"-"`
	IsLockingPosition bool    `json:"isLockingPosition" diff:"-"`
	IsLockingRotation bool    `json:"isLockingRotation" diff:"-"`
	IsHidingFrustum   bool    `json:"IsHidingFrustum" diff:"-"`
}
