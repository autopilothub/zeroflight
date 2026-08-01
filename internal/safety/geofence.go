package safety

import (
	"fmt"

	"github.com/autopilothub/zeroflight/internal/config"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/pkg/geo"
)

// Limits are geofence and altitude guard rails from configuration.
type Limits struct {
	MaxAltitudeM float32
	MaxRadiusM   float32
}

// LimitsFromConfig builds safety limits from file config.
func LimitsFromConfig(cfg config.Safety) Limits {
	return Limits{
		MaxAltitudeM: cfg.MaxAltitudeM,
		MaxRadiusM:   cfg.MaxRadiusM,
	}
}

// ValidateTarget checks a single target point against home and limits.
func ValidateTarget(home inav.HomePosition, lat, lon float64, altM float32, limits Limits) error {
	if limits.MaxAltitudeM > 0 && altM > limits.MaxAltitudeM {
		return fmt.Errorf("altitude %.1fm exceeds max %.1fm", altM, limits.MaxAltitudeM)
	}
	if limits.MaxRadiusM <= 0 {
		return nil
	}
	if !home.Valid {
		return fmt.Errorf("home position unknown; wait for GPS_GLOBAL_ORIGIN from INAV before autonomous commands")
	}
	dist := geo.DistanceM(home.Lat, home.Lon, lat, lon)
	if dist > float64(limits.MaxRadiusM) {
		return fmt.Errorf("target is %.0fm from home (max %.0fm)", dist, limits.MaxRadiusM)
	}
	return nil
}

// ValidateMission checks all mission waypoints against limits.
func ValidateMission(home inav.HomePosition, waypoints []mission.Waypoint, limits Limits) error {
	for i, wp := range waypoints {
		if err := ValidateTarget(home, wp.Lat, wp.Lon, wp.Alt, limits); err != nil {
			return fmt.Errorf("waypoint %d: %w", i, err)
		}
	}
	return nil
}
