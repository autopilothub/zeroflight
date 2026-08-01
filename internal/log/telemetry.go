package log

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
)

// TelemetryCSV appends vehicle state rows to a CSV file.
type TelemetryCSV struct {
	path    string
	headers bool
}

// NewTelemetryCSV creates a CSV logger at path.
func NewTelemetryCSV(path string) *TelemetryCSV {
	return &TelemetryCSV{path: path}
}

// Write records one telemetry snapshot.
func (l *TelemetryCSV) Write(state inav.VehicleState) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	exists := fileExists(l.path)
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open telemetry log: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if !exists {
		if err := w.Write([]string{
			"time", "lat", "lon", "rel_alt_m", "roll", "pitch", "yaw",
			"mode", "armed", "gcs_nav", "battery_v", "sats", "hdop",
		}); err != nil {
			return err
		}
	}

	ts := state.Time
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	armed := "0"
	if state.Armed {
		armed = "1"
	}
	gcsNav := "0"
	if state.GCSNavActive {
		gcsNav = "1"
	}

	if err := w.Write([]string{
		ts.Format(time.RFC3339Nano),
		fmt.Sprintf("%.7f", state.GPS.Lat),
		fmt.Sprintf("%.7f", state.GPS.Lon),
		fmt.Sprintf("%.2f", state.GPS.RelAltM),
		fmt.Sprintf("%.4f", state.Attitude.Roll),
		fmt.Sprintf("%.4f", state.Attitude.Pitch),
		fmt.Sprintf("%.4f", state.Attitude.Yaw),
		string(state.Mode),
		armed,
		gcsNav,
		fmt.Sprintf("%.2f", state.Battery.VoltageV),
		fmt.Sprintf("%d", state.GPS.Satellites),
		fmt.Sprintf("%.1f", state.GPS.HDOP),
	}); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
