package inav_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/autopilothub/zeroflight/internal/inav"
)

func TestGotoAuditLogger(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "goto.jsonl")
	logger := inav.NewGotoAuditLogger(path)

	entry := inav.GotoAuditEntry{
		Command: "hover",
		LatDeg:  37.5,
		LonDeg:  127.0,
		AltM:    12,
		Mode:    inav.ModeGCSNav,
		Armed:   true,
	}
	if err := logger.Log(entry); err != nil {
		t.Fatalf("log: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var decoded inav.GotoAuditEntry
	if err := json.Unmarshal(data[:len(data)-1], &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Command != "hover" || decoded.AltM != 12 {
		t.Fatalf("unexpected entry: %+v", decoded)
	}
}
