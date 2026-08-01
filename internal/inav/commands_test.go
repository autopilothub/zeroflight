package inav_test

import (
	"testing"

	"github.com/autopilothub/zeroflight/internal/inav"
)

func TestCheckGotoPreflightOK(t *testing.T) {
	state := inav.VehicleState{
		Connected:    true,
		Armed:        true,
		GCSNavActive: true,
		Mode:         inav.ModeGCSNav,
		GPS: inav.GPSFix{
			FixType:    3,
			Satellites: 10,
			HDOP:       1.2,
		},
	}
	_, err := inav.CheckGotoPreflight(state, inav.GotoPreflightOptions{})
	if err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestCheckGotoPreflightGCSNavHint(t *testing.T) {
	state := inav.VehicleState{
		Connected: true,
		Armed:     true,
		Mode:      inav.ModePosHold,
		GPS: inav.GPSFix{
			FixType:    3,
			Satellites: 10,
			HDOP:       1.0,
		},
	}
	_, err := inav.CheckGotoPreflight(state, inav.GotoPreflightOptions{})
	if err == nil {
		t.Fatal("expected error when GCS NAV inactive")
	}
}

func TestCheckGotoPreflightHDOPForce(t *testing.T) {
	state := inav.VehicleState{
		Connected:    true,
		Armed:        true,
		GCSNavActive: true,
		GPS: inav.GPSFix{
			FixType:    3,
			Satellites: 10,
			HDOP:       3.5,
		},
	}
	_, err := inav.CheckGotoPreflight(state, inav.GotoPreflightOptions{})
	if err == nil {
		t.Fatal("expected HDOP error without force")
	}

	result, err := inav.CheckGotoPreflight(state, inav.GotoPreflightOptions{Force: true})
	if err != nil {
		t.Fatalf("expected pass with force, got %v", err)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected HDOP warning with force")
	}
}
