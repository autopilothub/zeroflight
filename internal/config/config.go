package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"gopkg.in/yaml.v3"
)

// File is the top-level configuration loaded from YAML.
type File struct {
	Link    Link    `yaml:"link"`
	MAVLink MAVLink `yaml:"mavlink"`
	Safety  Safety  `yaml:"safety"`
	API     API     `yaml:"api"`
	MSP     MSP     `yaml:"msp"` // deprecated: use link section
}

// Link holds the primary FC serial connection settings.
type Link struct {
	Protocol string `yaml:"protocol"` // msp (default) or mavlink
	Device   string `yaml:"device"`
	Baud     int    `yaml:"baud"`
	PollHz   int    `yaml:"poll_hz"`
}

// MSP holds deprecated MSP settings (merged into link).
type MSP struct {
	Enabled bool   `yaml:"enabled"`
	Device  string `yaml:"device"`
	Baud    int    `yaml:"baud"`
	PollHz  int    `yaml:"poll_hz"`
}

// API holds HTTP server settings.
type API struct {
	Listen string `yaml:"listen"`
}

// MAVLink holds MAVLink-specific parameters (used when link.protocol=mavlink).
type MAVLink struct {
	Connection        string `yaml:"connection"`
	Device            string `yaml:"device"`
	Baud              int    `yaml:"baud"`
	TargetSystemID    uint8  `yaml:"target_system_id"`
	TargetComponentID uint8  `yaml:"target_component_id"`
	OutSystemID       uint8  `yaml:"out_system_id"`
	OutComponentID    uint8  `yaml:"out_component_id"`
	Version           int    `yaml:"version"`
}

// Safety holds navigation guard rails.
type Safety struct {
	MaxAltitudeM     float32 `yaml:"max_altitude_m"`
	MaxRadiusM       float32 `yaml:"max_radius_m"`
	ArrivalRadiusM   float64 `yaml:"arrival_radius_m"`
	ArrivalAltitudeM float32 `yaml:"arrival_altitude_m"`
	LinkTimeoutSec   float64 `yaml:"link_timeout_sec"`
}

// Default returns sensible defaults for Mamba F405 MK2 + INAV over MSP.
func Default() File {
	return File{
		Link: Link{
			Protocol: "msp",
			Device:   "/dev/serial0",
			Baud:     115200,
			PollHz:   10,
		},
		MAVLink: MAVLink{
			Connection:        "serial:/dev/serial0:115200",
			Device:            "/dev/serial0",
			Baud:              115200,
			TargetSystemID:    1,
			TargetComponentID: 1,
			OutSystemID:       255,
			OutComponentID:    190,
			Version:           2,
		},
		Safety: Safety{
			MaxAltitudeM:     120,
			MaxRadiusM:       500,
			ArrivalRadiusM:   3,
			ArrivalAltitudeM: 2,
			LinkTimeoutSec:   3,
		},
		API: API{
			Listen: "127.0.0.1:8080",
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
	cfg.normalize()
	return cfg, nil
}

func (f *File) normalize() {
	if f.Link.Protocol == "" {
		if f.MSP.Enabled || f.MSP.Device != "" {
			f.Link.Protocol = "msp"
		} else if f.MAVLink.Connection != "" {
			f.Link.Protocol = "mavlink"
		} else {
			f.Link.Protocol = "msp"
		}
	}
	if f.Link.Device == "" {
		if f.MSP.Device != "" {
			f.Link.Device = f.MSP.Device
		} else if f.MAVLink.Device != "" {
			f.Link.Device = f.MAVLink.Device
		}
	}
	if f.Link.Baud == 0 {
		if f.MSP.Baud > 0 {
			f.Link.Baud = f.MSP.Baud
		} else if f.MAVLink.Baud > 0 {
			f.Link.Baud = f.MAVLink.Baud
		}
	}
	if f.Link.PollHz == 0 && f.MSP.PollHz > 0 {
		f.Link.PollHz = f.MSP.PollHz
	}
}

// LinkConfig returns the active serial link settings.
func (f File) LinkConfig(connectionOverride string) Link {
	link := f.Link
	if connectionOverride != "" {
		if parsed, err := parseSerialConnection(connectionOverride); err == nil {
			link.Device = parsed.device
			link.Baud = parsed.baud
		}
	}
	if link.Device == "" {
		link.Device = "/dev/serial0"
	}
	if link.Baud == 0 {
		link.Baud = 115200
	}
	if link.PollHz == 0 {
		link.PollHz = 10
	}
	if link.Protocol == "" {
		link.Protocol = "msp"
	}
	return link
}

type serialConn struct {
	device string
	baud   int
}

func parseSerialConnection(raw string) (serialConn, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 || parts[0] != "serial" {
		return serialConn{}, fmt.Errorf("not a serial connection")
	}
	conn := serialConn{device: parts[1], baud: 115200}
	if len(parts) == 3 {
		if _, err := fmt.Sscanf(parts[2], "%d", &conn.baud); err != nil {
			return serialConn{}, err
		}
	}
	return conn, nil
}

// INAVConfig converts file settings into an INAV MAVLink client config.
func (f File) INAVConfig(connectionOverride string) (inav.Config, error) {
	base := inav.DefaultConfig()
	base.TargetSystemID = f.MAVLink.TargetSystemID
	base.TargetComponentID = f.MAVLink.TargetComponentID
	base.OutSystemID = f.MAVLink.OutSystemID
	base.OutComponentID = f.MAVLink.OutComponentID
	base.Device = f.MAVLink.Device
	base.Baud = f.MAVLink.Baud
	base.MAVLinkVersion = inav.MAVLinkVersion(f.MAVLink.Version)

	conn := connectionOverride
	if conn == "" {
		conn = f.MAVLink.Connection
	}
	if conn == "" {
		link := f.LinkConfig("")
		base.Device = link.Device
		base.Baud = link.Baud
		return base, nil
	}
	return inav.ParseConnection(conn, base)
}

// LinkTimeout returns the FC stale link threshold.
func (f File) LinkTimeout() time.Duration {
	if f.Safety.LinkTimeoutSec <= 0 {
		return 3 * time.Second
	}
	return time.Duration(f.Safety.LinkTimeoutSec * float64(time.Second))
}

// ListenAddr returns the HTTP API listen address.
func (f File) ListenAddr(override string) string {
	if override != "" {
		return override
	}
	if f.API.Listen != "" {
		return f.API.Listen
	}
	return "127.0.0.1:8080"
}
