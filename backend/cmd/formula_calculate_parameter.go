package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type CameraParams struct {
	Vendor      string  `json:"vendor" yaml:"vendor"`
	Camera      string  `json:"camera" yaml:"camera"`
	SensorName  string  `json:"sensor_name" yaml:"sensor_name"`
	Aspect      float64 `json:"aspect" yaml:"aspect"`
	FOV         float64 `json:"fov" yaml:"fov"`
	Near        float64 `json:"near" yaml:"near"`
	Far         float64 `json:"far" yaml:"far"`
	PixelPitch  float64 `json:"pixel_pitch" yaml:"pixel_pitch"`
	ResolutionW int     `json:"resolution_width" yaml:"resolution_width"`
	ResolutionH int     `json:"resolution_height" yaml:"resolution_height"`
	SensorWmm   float64 `json:"sensor_width_mm" yaml:"sensor_width_mm"`
	SensorHmm   float64 `json:"sensor_height_mm" yaml:"sensor_height_mm"`
	FocalLength float64 `json:"focal_length" yaml:"focal_length"`
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func main() {
	url := "https://raw.githubusercontent.com/EmberLightVFX/Camera-Sensor-Database/main/data/sensors.csv"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	var results []CameraParams

	// ข้าม header แถวแรก
	for i, row := range records {
		if i == 0 {
			continue
		}

		vendor := row[0]
		camera := row[1]
		sensorName := row[2]

		focalLength := parseFloat(row[3], 35.0)
		resW := parseInt(row[4], 0)
		resH := parseInt(row[5], 0)
		sensorWmm := parseFloat(row[6], 0)
		sensorHmm := parseFloat(row[7], 0)

		if resW == 0 || resH == 0 || sensorWmm == 0 || sensorHmm == 0 {
			continue
		}

		aspect := float64(resW) / float64(resH)
		fov := 2 * math.Atan((sensorHmm/2.0)/focalLength) * (180.0 / math.Pi)
		pixelPitch := sensorWmm / float64(resW)

		results = append(results, CameraParams{
			Vendor:      vendor,
			Camera:      camera,
			SensorName:  sensorName,
			Aspect:      aspect,
			FOV:         fov,
			Near:        0.1,
			Far:         1000,
			PixelPitch:  pixelPitch,
			ResolutionW: resW,
			ResolutionH: resH,
			SensorWmm:   sensorWmm,
			SensorHmm:   sensorHmm,
			FocalLength: focalLength,
		})
	}

	// save JSON
	jf, _ := os.Create("../OmniCam/backend/data/camera_parameter.json")
	defer jf.Close()
	jEnc := json.NewEncoder(jf)
	jEnc.SetIndent("", "  ")
	jEnc.Encode(results)

	// save YAML
	yf, _ := os.Create("../OmniCam/backend/data/camera_parameter.yaml")
	defer yf.Close()
	yEnc := yaml.NewEncoder(yf)
	yEnc.Encode(results)

	// save CSV
	cf, _ := os.Create("../OmniCam/backend/data/camera_parameter.csv")
	defer cf.Close()
	writer := csv.NewWriter(cf)
	defer writer.Flush()

	writer.Write([]string{"vendor", "camera", "sensor_name", "aspect", "fov", "near", "far", "pixel_pitch", "res_w", "res_h", "sensor_w_mm", "sensor_h_mm", "focal_length"})
	for _, r := range results {
		writer.Write([]string{
			r.Vendor,
			r.Camera,
			r.SensorName,
			fmt.Sprintf("%.4f", r.Aspect),
			fmt.Sprintf("%.2f", r.FOV),
			fmt.Sprintf("%.2f", r.Near),
			fmt.Sprintf("%.2f", r.Far),
			fmt.Sprintf("%.6f", r.PixelPitch),
			fmt.Sprintf("%d", r.ResolutionW),
			fmt.Sprintf("%d", r.ResolutionH),
			fmt.Sprintf("%.2f", r.SensorWmm),
			fmt.Sprintf("%.2f", r.SensorHmm),
			fmt.Sprintf("%.2f", r.FocalLength),
		})
	}

	fmt.Println("Exported: camera_parameter.json, camera_parameter.yaml, camera_parameter.csv")
}
