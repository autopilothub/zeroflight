package inav

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const defaultGotoLogPath = "logs/goto.jsonl"

// GotoAuditEntry records a goto or hover command for post-flight analysis.
type GotoAuditEntry struct {
	Time      time.Time  `json:"time"`
	Command   string     `json:"command"`
	LatDeg    float64    `json:"lat_deg"`
	LonDeg    float64    `json:"lon_deg"`
	AltM      float32    `json:"alt_m"`
	Mode      FlightMode `json:"mode"`
	Armed     bool       `json:"armed"`
	FixType   uint8      `json:"fix_type"`
	Satellites uint8     `json:"satellites"`
	HDOP      float32    `json:"hdop"`
	Forced    bool       `json:"forced"`
}

// GotoAuditLogger appends goto/hover events as JSON lines.
type GotoAuditLogger struct {
	path string
	mu   sync.Mutex
}

// NewGotoAuditLogger creates a logger writing to path (default logs/goto.jsonl).
func NewGotoAuditLogger(path string) *GotoAuditLogger {
	if path == "" {
		path = defaultGotoLogPath
	}
	return &GotoAuditLogger{path: path}
}

// Log appends one audit entry to the log file.
func (l *GotoAuditLogger) Log(entry GotoAuditEntry) error {
	if entry.Time.IsZero() {
		entry.Time = time.Now().UTC()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open goto log: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal goto log entry: %w", err)
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write goto log: %w", err)
	}
	return nil
}
