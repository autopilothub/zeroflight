package config

import (
	"fmt"
	"os"

	"github.com/autopilothub/zeroflight/internal/inav"
	"gopkg.in/yaml.v3"
)

// File is the top-level configuration loaded from YAML.
type File struct {
	MAVLink MAVLink `yaml:"mavlink"`
	Safety  Safety  `yaml:"safety"`
}

// MAVLink holds connection parameters.
type MAVLink struct {
	Connection        string `yaml:"connection"`
	Device            string `yaml:"device"`
	Baud              int    `yaml:"baud"`
	TargetSystemID    uint8  `yaml:"target_system_id"`
	TargetComponentID uint8  `yaml:"target_component_id"`
	OutSystemID       uint8  `yaml:"out_system_id"`
	OutComponentID    uint8  `yaml:"out_component_id"`
}

// Safety holds navigation guard rails.
type Safety struct {
	MaxAltitudeM      float32 `yaml:"max_altitude_m"`
	MaxRadiusM        float32 `yaml:"max_radius_m"`
	ArrivalRadiusM    float64 `yaml:"arrival_radius_m"`
	ArrivalAltitudeM  float32 `yaml:"arrival_altitude_m"`
}

// Default returns sensible defaults for Mamba F405 MK2 + INAV.
func Default() File {
	return File{
		MAVLink: MAVLink{
			Connection:        "serial:/dev/ttyAMA0:115200",
			Device:            "/dev/ttyAMA0",
			Baud:              115200,
			TargetSystemID:    1,
			TargetComponentID: 1,
			OutSystemID:       255,
			OutComponentID:    190,
		},
		Safety: Safety{
			MaxAltitudeM:     120,
			MaxRadiusM:       500,
			ArrivalRadiusM:   3,
			ArrivalAltitudeM: 2,
		},
	}
}

// Load reads a YAML config file, falling back to defaults for missing fields.
func Load(path string) (File, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// INAVConfig converts file settings into an INAV client config.
func (f File) INAVConfig(connectionOverride string) (inav.Config, error) {
	base := inav.DefaultConfig()
	base.TargetSystemID = f.MAVLink.TargetSystemID
	base.TargetComponentID = f.MAVLink.TargetComponentID
	base.OutSystemID = f.MAVLink.OutSystemID
	base.OutComponentID = f.MAVLink.OutComponentID
	base.Device = f.MAVLink.Device
	base.Baud = f.MAVLink.Baud

	conn := connectionOverride
	if conn == "" {
		conn = f.MAVLink.Connection
	}
	if conn == "" {
		return base, nil
	}
	return inav.ParseConnection(conn, base)
}
