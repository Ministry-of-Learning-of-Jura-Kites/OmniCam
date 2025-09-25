package messages_cameras

type Cameras = map[string]CameraStruct

type CameraStruct struct {
	Name   string  `json:"name" binding:"required"`
	AngleX float64 `json:"angleX"`
	AngleY float64 `json:"angleY"`
	AngleZ float64 `json:"angleZ"`
	AngleW float64 `json:"angleW"`
	PosX   float64 `json:"posX"`
	PosY   float64 `json:"posY"`
	PosZ   float64 `json:"posZ"`
	Fov    float64 `json:"fov"`
}
