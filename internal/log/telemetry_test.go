package log_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	telemetrylog "github.com/autopilothub/zeroflight/internal/log"
	"github.com/autopilothub/zeroflight/internal/inav"
)

func TestTelemetryCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "telemetry.csv")
	logger := telemetrylog.NewTelemetryCSV(path)

	state := inav.VehicleState{
		Time:         time.Now().UTC(),
		Mode:         inav.ModeGCSNav,
		Armed:        true,
		GCSNavActive: true,
		GPS:          inav.GPSFix{Lat: 37.5, Lon: 127.0, RelAltM: 12, Satellites: 10, HDOP: 1.1},
		Battery:      inav.Battery{VoltageV: 16.2},
	}
	if err := logger.Write(state); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected csv content")
	}
}
