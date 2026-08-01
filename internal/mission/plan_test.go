package mission_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/autopilothub/zeroflight/internal/mission"
)

func TestLoadPlan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mission.yaml")
	content := `waypoints:
  - lat: 37.5665
    lon: 126.9780
    alt: 15
  - lat: 37.5670
    lon: 126.9785
    alt: 20
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	plan, err := mission.Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(plan.Waypoints) != 2 {
		t.Fatalf("expected 2 waypoints, got %d", len(plan.Waypoints))
	}
}

func TestValidateEmpty(t *testing.T) {
	err := (mission.Plan{}).Validate()
	if err == nil {
		t.Fatal("expected error for empty mission")
	}
}
