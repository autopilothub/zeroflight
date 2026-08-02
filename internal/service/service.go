package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/autopilothub/zeroflight/internal/config"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/safety"
	"github.com/autopilothub/zeroflight/internal/session"
	"github.com/autopilothub/zeroflight/pkg/geo"
)

// Service holds a long-lived MAVLink session for API and serve mode.
type Service struct {
	session    *session.Session
	gotoLogger *inav.GotoAuditLogger
}

// New connects to the flight controller and starts background telemetry.
func New(ctx context.Context, cfgPath, connectionOverride string) (*Service, error) {
	sess, err := session.Open(ctx, cfgPath, connectionOverride)
	if err != nil {
		return nil, err
	}
	return &Service{
		session:    sess,
		gotoLogger: inav.NewGotoAuditLogger(""),
	}, nil
}

// Close shuts down connections.
func (s *Service) Close() {
	s.session.Close()
}

// WaitReady blocks until telemetry is available.
func (s *Service) WaitReady(ctx context.Context) error {
	return s.session.WaitReady(ctx)
}

// State returns the latest vehicle snapshot.
func (s *Service) State() inav.VehicleState {
	return s.session.State()
}

// Config returns the loaded configuration.
func (s *Service) Config() config.File {
	return s.session.Config()
}

// Preflight runs the autonomous flight checklist.
func (s *Service) Preflight() (bool, []safety.CheckResult) {
	limits := safety.LimitsFromConfig(s.Config().Safety)
	results := safety.RunPreflight(s.State(), limits, s.Config().LinkTimeout())
	return safety.AllPassed(results), results
}

// GotoOptions configures a goto/hover command.
type GotoOptions struct {
	Command string
	Lat     float64
	Lon     float64
	Alt     float32
	Yaw     *float32
	Force   bool
	Hover   bool
}

// Goto sends DO_REPOSITION after safety checks.
func (s *Service) Goto(ctx context.Context, opts GotoOptions) error {
	state := s.State()
	if safety.IsLinkStale(state, s.Config().LinkTimeout()) {
		return fmt.Errorf("msp link stale")
	}

	preflight, err := inav.CheckGotoPreflight(state, inav.GotoPreflightOptions{Force: opts.Force})
	if err != nil {
		return err
	}
	_ = preflight

	lat, lon, alt := opts.Lat, opts.Lon, opts.Alt
	if opts.Hover {
		if state.GPS.Lat == 0 && state.GPS.Lon == 0 {
			return fmt.Errorf("no valid GPS position for hover")
		}
		lat = state.GPS.Lat
		lon = state.GPS.Lon
		if alt == 0 {
			alt = state.GPS.RelAltM
		}
	}

	limits := safety.LimitsFromConfig(s.Config().Safety)
	if err := safety.ValidateTarget(state.Home, lat, lon, alt, limits); err != nil {
		return err
	}

	req := inav.GotoRequest{LatDeg: lat, LonDeg: lon, AltM: alt, YawDeg: opts.Yaw}
	if err := s.session.SendGoto(req); err != nil {
		return err
	}

	command := opts.Command
	if command == "" {
		command = "goto"
	}
	_ = s.gotoLogger.Log(inav.GotoAuditEntry{
		Command:    command,
		LatDeg:     lat,
		LonDeg:     lon,
		AltM:       alt,
		Mode:       state.Mode,
		Armed:      state.Armed,
		FixType:    state.GPS.FixType,
		Satellites: state.GPS.Satellites,
		HDOP:       state.GPS.HDOP,
		Forced:     opts.Force,
	})
	return nil
}

// UploadMission uploads waypoints when disarmed.
func (s *Service) UploadMission(ctx context.Context, waypoints []mission.Waypoint) error {
	state := s.State()
	if state.Armed {
		return fmt.Errorf("disarm the vehicle before uploading a mission")
	}
	limits := safety.LimitsFromConfig(s.Config().Safety)
	if err := safety.ValidateMission(state.Home, waypoints, limits); err != nil {
		return err
	}
	return s.session.UploadMission(ctx, waypoints)
}

// ClearMission removes stored mission items when disarmed.
func (s *Service) ClearMission(ctx context.Context) error {
	if s.State().Armed {
		return fmt.Errorf("disarm the vehicle before clearing a mission")
	}
	return s.session.ClearMission(ctx)
}

// WaitArrival blocks until the vehicle reaches a target or times out.
func (s *Service) WaitArrival(ctx context.Context, lat, lon float64, alt float32, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	cfg := s.Config()
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
		cur := s.State()
		if safety.IsLinkStale(cur, cfg.LinkTimeout()) {
			return fmt.Errorf("fc link lost during wait")
		}
		if !cur.GCSNavActive {
			return fmt.Errorf("GCS NAV deactivated during wait")
		}
		horiz := geo.DistanceM(cur.GPS.Lat, cur.GPS.Lon, lat, lon)
		altErr := float32(math.Abs(float64(cur.GPS.RelAltM - alt)))
		if horiz <= cfg.Safety.ArrivalRadiusM && altErr <= cfg.Safety.ArrivalAltitudeM {
			return nil
		}
	}
	return fmt.Errorf("timed out waiting for arrival")
}
