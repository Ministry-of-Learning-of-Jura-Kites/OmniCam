package messages_optimization

// TODO: Use protobufs instead
type OptimizationCamera struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	AngleX    float64 `json:"angle_x"`
	AngleY    float64 `json:"angle_y"`
	AngleZ    float64 `json:"angle_z"`
	AngleW    float64 `json:"angle_w"`
	PosX      float64 `json:"pos_x"`
	PosY      float64 `json:"pos_y"`
	PosZ      float64 `json:"pos_z"`
	Fov       float64 `json:"fov"`
	WidthRes  int     `json:"width_res"`
	HeightRes int     `json:"height_res"`
}

type OptimizeResponse struct {
	Cameras []OptimizationCamera `json:"cameras"`
	JobId   string               `json:"job_id"`
}

