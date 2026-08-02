package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/safety"
	"github.com/autopilothub/zeroflight/internal/service"
	"github.com/autopilothub/zeroflight/internal/web"
)

// Controller is the API surface used by HTTP handlers.
type Controller interface {
	State() inav.VehicleState
	Preflight() (bool, []safety.CheckResult)
	Goto(ctx context.Context, opts service.GotoOptions) error
	UploadMission(ctx context.Context, waypoints []mission.Waypoint) error
	ClearMission(ctx context.Context) error
}

// Server exposes ZeroFlight over HTTP.
type Server struct {
	svc Controller
	mux *http.ServeMux
}

// NewServer creates an API server for the given controller.
func NewServer(svc Controller) *Server {
	s := &Server{svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/v1/preflight", s.handlePreflight)
	s.mux.HandleFunc("POST /api/v1/goto", s.handleGoto)
	s.mux.HandleFunc("POST /api/v1/hover", s.handleHover)
	s.mux.HandleFunc("POST /api/v1/mission/upload", s.handleMissionUpload)
	s.mux.HandleFunc("POST /api/v1/mission/clear", s.handleMissionClear)
	if err := web.Register(s.mux); err != nil {
		panic(err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, toVehicleJSON(s.svc.State()))
}

func (s *Server) handlePreflight(w http.ResponseWriter, r *http.Request) {
	passed, results := s.svc.Preflight()
	checks := make([]checkJSON, len(results))
	for i, c := range results {
		checks[i] = checkJSON{Name: c.Name, Passed: c.Passed, Message: c.Message}
	}
	writeJSON(w, http.StatusOK, preflightJSON{Passed: passed, Checks: checks})
}

func (s *Server) handleGoto(w http.ResponseWriter, r *http.Request) {
	var req gotoRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	opts := service.GotoOptions{
		Command: "goto",
		Lat:     req.Lat,
		Lon:     req.Lon,
		Alt:     req.Alt,
		Force:   req.Force,
	}
	if req.SetYaw {
		opts.Yaw = &req.Yaw
	}
	if err := s.svc.Goto(r.Context(), opts); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (s *Server) handleHover(w http.ResponseWriter, r *http.Request) {
	var req hoverRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.Goto(r.Context(), service.GotoOptions{
		Command: "hover",
		Alt:     req.Alt,
		Force:   req.Force,
		Hover:   true,
	}); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (s *Server) handleMissionUpload(w http.ResponseWriter, r *http.Request) {
	var req missionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	plan := mission.Plan{Waypoints: req.Waypoints}
	if err := plan.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.UploadMission(r.Context(), plan.Waypoints); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "uploaded",
		"waypoints":  len(plan.Waypoints),
	})
}

func (s *Server) handleMissionClear(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.ClearMission(r.Context()); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

type gotoRequest struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Alt    float32 `json:"alt"`
	Yaw    float32 `json:"yaw"`
	SetYaw bool    `json:"set_yaw"`
	Force  bool    `json:"force"`
}

type hoverRequest struct {
	Alt   float32 `json:"alt"`
	Force bool    `json:"force"`
}

type missionRequest struct {
	Waypoints []mission.Waypoint `json:"waypoints"`
}

type checkJSON struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type preflightJSON struct {
	Passed bool        `json:"passed"`
	Checks []checkJSON `json:"checks"`
}

type vehicleJSON struct {
	Time         time.Time      `json:"time"`
	LinkOpen     bool           `json:"link_open"`
	Connected    bool           `json:"connected"`
	ParseErrors  uint64         `json:"parse_errors"`
	Armed        bool           `json:"armed"`
	Mode         inav.FlightMode `json:"mode"`
	GCSNavActive bool           `json:"gcs_nav_active"`
	Attitude     attitudeJSON   `json:"attitude"`
	GPS          gpsJSON        `json:"gps"`
	Battery      batteryJSON    `json:"battery"`
	Sensors      sensorsJSON    `json:"sensors"`
	Home         homeJSON       `json:"home"`
	RawIMU       rawIMUJSON     `json:"raw_imu"`
}

type attitudeJSON struct {
	Roll       float32 `json:"roll"`
	Pitch      float32 `json:"pitch"`
	Yaw        float32 `json:"yaw"`
	RollSpeed  float32 `json:"roll_speed"`
	PitchSpeed float32 `json:"pitch_speed"`
	YawSpeed   float32 `json:"yaw_speed"`
}

type gpsJSON struct {
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	AltM        float32 `json:"alt_m"`
	RelAltM     float32 `json:"rel_alt_m"`
	FixType     uint8   `json:"fix_type"`
	Satellites  uint8   `json:"satellites"`
	HDOP        float32 `json:"hdop"`
	GroundSpeed float32 `json:"ground_speed"`
	ClimbRate   float32 `json:"climb_rate"`
}

type batteryJSON struct {
	VoltageV     float32 `json:"voltage_v"`
	CurrentA     float32 `json:"current_a"`
	RemainingPct int8    `json:"remaining_pct"`
}

type sensorsJSON struct {
	Gyro  bool `json:"gyro"`
	Accel bool `json:"accel"`
	Mag   bool `json:"mag"`
	Baro  bool `json:"baro"`
	GPS   bool `json:"gps"`
	RC    bool `json:"rc"`
}

type homeJSON struct {
	Valid bool    `json:"valid"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	AltM  float32 `json:"alt_m"`
}

type rawIMUJSON struct {
	Available bool      `json:"available"`
	Time      time.Time `json:"time"`
	Accel     [3]int16  `json:"accel"`
	Gyro      [3]int16  `json:"gyro"`
	Mag       [3]int16  `json:"mag"`
}

func toVehicleJSON(state inav.VehicleState) vehicleJSON {
	return vehicleJSON{
		Time:         state.Time,
		LinkOpen:     state.LinkOpen,
		Connected:    state.Connected,
		ParseErrors:  state.ParseErrors,
		Armed:        state.Armed,
		Mode:         state.Mode,
		GCSNavActive: state.GCSNavActive,
		Attitude: attitudeJSON{
			Roll: state.Attitude.Roll, Pitch: state.Attitude.Pitch, Yaw: state.Attitude.Yaw,
			RollSpeed: state.Attitude.RollSpeed, PitchSpeed: state.Attitude.PitchSpeed, YawSpeed: state.Attitude.YawSpeed,
		},
		GPS: gpsJSON{
			Lat: state.GPS.Lat, Lon: state.GPS.Lon, AltM: state.GPS.AltM, RelAltM: state.GPS.RelAltM,
			FixType: state.GPS.FixType, Satellites: state.GPS.Satellites, HDOP: state.GPS.HDOP,
			GroundSpeed: state.GPS.GroundSpeed, ClimbRate: state.GPS.ClimbRate,
		},
		Battery: batteryJSON{
			VoltageV: state.Battery.VoltageV, CurrentA: state.Battery.CurrentA, RemainingPct: state.Battery.RemainingPct,
		},
		Sensors: sensorsJSON{
			Gyro: state.Sensors.Gyro, Accel: state.Sensors.Accel, Mag: state.Sensors.Mag,
			Baro: state.Sensors.Baro, GPS: state.Sensors.GPS, RC: state.Sensors.RC,
		},
		Home: homeJSON{
			Valid: state.Home.Valid, Lat: state.Home.Lat, Lon: state.Home.Lon, AltM: state.Home.AltM,
		},
		RawIMU: rawIMUJSON{
			Available: state.RawIMU.Available,
			Time:      state.RawIMU.Time,
			Accel:     state.RawIMU.Accel,
			Gyro:      state.RawIMU.Gyro,
			Mag:       state.RawIMU.Mag,
		},
	}
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return errors.New("empty request body")
	}
	return json.Unmarshal(body, dst)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
