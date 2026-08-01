package inav_test

import (
	"testing"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

func TestParseCopterMode(t *testing.T) {
	mode, ok := inav.ParseCopterMode(uint32(ardupilotmega.COPTER_MODE_GUIDED))
	if !ok || mode != inav.ModeGCSNav {
		t.Fatalf("expected GCS_NAV, got %s ok=%v", mode, ok)
	}

	mode, ok = inav.ParseCopterMode(uint32(ardupilotmega.COPTER_MODE_LOITER))
	if !ok || mode != inav.ModePosHold {
		t.Fatalf("expected POS_HOLD, got %s ok=%v", mode, ok)
	}
}
