package safety

import (
	"fmt"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/pkg/geo"
)

// CheckResult is one preflight checklist line.
type CheckResult struct {
	Name    string
	Passed  bool
	Message string
}

// RunPreflight builds a checklist for autonomous flight readiness.
func RunPreflight(state inav.VehicleState, limits Limits, linkTimeout time.Duration) []CheckResult {
	var results []CheckResult

	results = append(results, check("FC link", state.Connected && !IsLinkStale(state, linkTimeout),
		linkMessage(state, linkTimeout)))

	results = append(results, check("Armed", state.Armed, armedMessage(state)))
	results = append(results, check("GPS 3D fix", state.GPS.FixType >= 3, gpsFixMessage(state)))
	results = append(results, check("Satellites >= 6", state.GPS.Satellites >= 6, satMessage(state)))
	results = append(results, check("GCS NAV active", state.GCSNavActive, gcsNavMessage(state)))
	results = append(results, check("Sensors healthy",
		state.Sensors.Gyro && state.Sensors.Accel && state.Sensors.GPS && state.Sensors.Baro,
		sensorMessage(state)))

	if limits.MaxRadiusM > 0 {
		results = append(results, check("Home position known", state.Home.Valid, homeMessage(state)))
		if state.Home.Valid && state.GPS.FixType >= 3 {
			dist := distanceFromHome(state)
			results = append(results, check(fmt.Sprintf("Within geofence (%.0fm)", limits.MaxRadiusM),
				dist <= float64(limits.MaxRadiusM),
				fmt.Sprintf("current distance from home: %.0fm", dist)))
		}
	}

	if limits.MaxAltitudeM > 0 && state.GPS.FixType >= 3 {
		results = append(results, check(fmt.Sprintf("Altitude <= %.0fm", limits.MaxAltitudeM),
			state.GPS.RelAltM <= limits.MaxAltitudeM,
			fmt.Sprintf("current relative altitude: %.1fm", state.GPS.RelAltM)))
	}

	return results
}

// AllPassed reports whether every checklist item passed.
func AllPassed(results []CheckResult) bool {
	for _, r := range results {
		if !r.Passed {
			return false
		}
	}
	return true
}

// IsLinkStale returns true when telemetry is older than timeout.
func IsLinkStale(state inav.VehicleState, timeout time.Duration) bool {
	if timeout <= 0 || state.Time.IsZero() {
		return false
	}
	return time.Since(state.Time) > timeout
}

func check(name string, passed bool, message string) CheckResult {
	return CheckResult{Name: name, Passed: passed, Message: message}
}

func linkMessage(state inav.VehicleState, timeout time.Duration) string {
	if !state.Connected {
		return "not connected"
	}
	if IsLinkStale(state, timeout) {
		return fmt.Sprintf("stale telemetry (last update %s ago)", time.Since(state.Time).Round(time.Second))
	}
	return "ok"
}

func armedMessage(state inav.VehicleState) string {
	if state.Armed {
		return "armed"
	}
	return "disarmed"
}

func gpsFixMessage(state inav.VehicleState) string {
	return fmt.Sprintf("fix type %d", state.GPS.FixType)
}

func satMessage(state inav.VehicleState) string {
	return fmt.Sprintf("%d satellites", state.GPS.Satellites)
}

func gcsNavMessage(state inav.VehicleState) string {
	if state.GCSNavActive {
		return "GCS NAV on"
	}
	if state.Mode == inav.ModePosHold {
		return "in POS HOLD — enable GCS NAV switch"
	}
	return fmt.Sprintf("mode %s — enable GCS NAV", state.Mode)
}

func sensorMessage(state inav.VehicleState) string {
	return fmt.Sprintf("gyro=%v accel=%v baro=%v gps=%v",
		state.Sensors.Gyro, state.Sensors.Accel, state.Sensors.Baro, state.Sensors.GPS)
}

func homeMessage(state inav.VehicleState) string {
	if state.Home.Valid {
		return fmt.Sprintf("home %.7f, %.7f", state.Home.Lat, state.Home.Lon)
	}
	return "waiting for GPS_GLOBAL_ORIGIN from INAV"
}

func distanceFromHome(state inav.VehicleState) float64 {
	if !state.Home.Valid {
		return 0
	}
	return geo.DistanceM(state.Home.Lat, state.Home.Lon, state.GPS.Lat, state.GPS.Lon)
}
