package safety_test

import (
	"testing"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/safety"
)

func TestValidateTargetGeofence(t *testing.T) {
	home := inav.HomePosition{Lat: 37.0, Lon: 127.0, Valid: true}
	limits := safety.Limits{MaxAltitudeM: 50, MaxRadiusM: 100}

	if err := safety.ValidateTarget(home, 37.0005, 127.0, 20, limits); err != nil {
		t.Fatalf("expected inside geofence: %v", err)
	}

	if err := safety.ValidateTarget(home, 38.0, 127.0, 20, limits); err == nil {
		t.Fatal("expected outside geofence error")
	}
}

func TestValidateTargetNoHome(t *testing.T) {
	limits := safety.Limits{MaxRadiusM: 100}
	err := safety.ValidateTarget(inav.HomePosition{}, 37.0, 127.0, 10, limits)
	if err == nil {
		t.Fatal("expected home unknown error")
	}
}

func TestValidateMission(t *testing.T) {
	home := inav.HomePosition{Lat: 37.0, Lon: 127.0, Valid: true}
	limits := safety.Limits{MaxAltitudeM: 50, MaxRadiusM: 500}
	plan := []mission.Waypoint{
		{Lat: 37.001, Lon: 127.0, Alt: 15},
		{Lat: 37.002, Lon: 127.0, Alt: 15},
	}
	if err := safety.ValidateMission(home, plan, limits); err != nil {
		t.Fatalf("mission should pass: %v", err)
	}
}

func TestIsLinkStale(t *testing.T) {
	state := inav.VehicleState{Time: time.Now().Add(-5 * time.Second)}
	if !safety.IsLinkStale(state, 2*time.Second) {
		t.Fatal("expected stale link")
	}
}

func TestRunPreflight(t *testing.T) {
	state := inav.VehicleState{
		Connected:    true,
		Time:         time.Now(),
		Armed:        true,
		GCSNavActive: true,
		Mode:         inav.ModeGCSNav,
		GPS:          inav.GPSFix{FixType: 3, Satellites: 10, RelAltM: 12, Lat: 37.0005, Lon: 127.0},
		Sensors:      inav.SensorHealth{Gyro: true, Accel: true, Baro: true, GPS: true},
		Home:         inav.HomePosition{Lat: 37.0, Lon: 127.0, Valid: true},
	}
	results := safety.RunPreflight(state, safety.Limits{MaxRadiusM: 500, MaxAltitudeM: 120}, 3*time.Second)
	if !safety.AllPassed(results) {
		t.Fatalf("expected all passed: %+v", results)
	}
}
