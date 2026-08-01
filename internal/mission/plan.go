package mission

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Waypoint is a single mission position in WGS84.
type Waypoint struct {
	Lat float64 `yaml:"lat"`
	Lon float64 `yaml:"lon"`
	Alt float32 `yaml:"alt"`
}

// Plan is a YAML mission file with ordered waypoints.
type Plan struct {
	Waypoints []Waypoint `yaml:"waypoints"`
}

// Load reads a mission plan from a YAML file.
func Load(path string) (Plan, error) {
	var plan Plan
	data, err := os.ReadFile(path)
	if err != nil {
		return plan, fmt.Errorf("read mission file: %w", err)
	}
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return plan, fmt.Errorf("parse mission file: %w", err)
	}
	if err := plan.Validate(); err != nil {
		return plan, err
	}
	return plan, nil
}

// Validate checks mission plan constraints.
func (p Plan) Validate() error {
	if len(p.Waypoints) == 0 {
		return fmt.Errorf("mission must contain at least one waypoint")
	}
	for i, wp := range p.Waypoints {
		if wp.Lat < -90 || wp.Lat > 90 {
			return fmt.Errorf("waypoint %d: invalid latitude %.7f", i, wp.Lat)
		}
		if wp.Lon < -180 || wp.Lon > 180 {
			return fmt.Errorf("waypoint %d: invalid longitude %.7f", i, wp.Lon)
		}
		if wp.Alt < 0 {
			return fmt.Errorf("waypoint %d: altitude must be >= 0", i)
		}
	}
	return nil
}
