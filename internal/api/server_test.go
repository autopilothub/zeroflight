package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/autopilothub/zeroflight/internal/api"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/safety"
	"github.com/autopilothub/zeroflight/internal/service"
)

type stubController struct {
	state inav.VehicleState
}

func (s *stubController) State() inav.VehicleState { return s.state }
func (s *stubController) Preflight() (bool, []safety.CheckResult) {
	return true, []safety.CheckResult{{Name: "test", Passed: true, Message: "ok"}}
}
func (s *stubController) Goto(context.Context, service.GotoOptions) error { return nil }
func (s *stubController) UploadMission(context.Context, []mission.Waypoint) error { return nil }
func (s *stubController) ClearMission(context.Context) error { return nil }

func TestHealthEndpoint(t *testing.T) {
	srv := api.NewServer(&stubController{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestStatusEndpoint(t *testing.T) {
	now := time.Now()
	srv := api.NewServer(&stubController{state: inav.VehicleState{
		Time: now, Connected: true, Mode: inav.ModeGCSNav, Armed: true,
	}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["mode"] != string(inav.ModeGCSNav) {
		t.Fatalf("unexpected mode %v", body["mode"])
	}
}

func TestGotoEndpoint(t *testing.T) {
	srv := api.NewServer(&stubController{})
	body := `{"lat":37.5,"lon":127.0,"alt":15}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/goto", strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestDashboardIndex(t *testing.T) {
	srv := api.NewServer(&stubController{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ZeroFlight GCS") {
		t.Fatalf("expected dashboard HTML")
	}
}

func TestDashboardAssets(t *testing.T) {
	srv := api.NewServer(&stubController{})
	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "pollStatus") {
		t.Fatalf("expected app.js content")
	}
}
