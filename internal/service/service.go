package service

import (
	"context"
	"fmt"
	"time"

	"github.com/autopilothub/zeroflight/internal/config"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/safety"
)

// Service holds a long-lived MAVLink session for API and serve mode.
type Service struct {
	cfg        config.File
	client     *inav.Client
	cancel     context.CancelFunc
	gotoLogger *inav.GotoAuditLogger
}

// New connects to the flight controller and starts background telemetry.
func New(ctx context.Context, cfgPath, connectionOverride string) (*Service, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	clientCfg, err := cfg.INAVConfig(connectionOverride)
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctx)
	client := inav.NewClient(clientCfg)
	if err := client.Connect(runCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("connect: %w", err)
	}

	s := &Service{
		cfg:        cfg,
		client:     client,
		cancel:     cancel,
		gotoLogger: inav.NewGotoAuditLogger(""),
	}
	return s, nil
}

// Close shuts down the MAVLink client.
func (s *Service) Close() {
	s.cancel()
	s.client.Close()
}

// WaitReady blocks until telemetry is available.
func (s *Service) WaitReady(ctx context.Context) error {
	return s.client.WaitForConnection(ctx, 15*time.Second)
}

// State returns the latest vehicle snapshot.
func (s *Service) State() inav.VehicleState {
	return s.client.State()
}

// Config returns the loaded configuration.
func (s *Service) Config() config.File {
	return s.cfg
}

// Preflight runs the autonomous flight checklist.
func (s *Service) Preflight() (bool, []safety.CheckResult) {
	limits := safety.LimitsFromConfig(s.cfg.Safety)
	results := safety.RunPreflight(s.State(), limits, s.cfg.LinkTimeout())
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
	if safety.IsLinkStale(state, s.cfg.LinkTimeout()) {
		return fmt.Errorf("mavlink link stale")
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

	limits := safety.LimitsFromConfig(s.cfg.Safety)
	if err := safety.ValidateTarget(state.Home, lat, lon, alt, limits); err != nil {
		return err
	}

	req := inav.GotoRequest{LatDeg: lat, LonDeg: lon, AltM: alt, YawDeg: opts.Yaw}
	if err := s.client.SendGoto(req); err != nil {
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
	limits := safety.LimitsFromConfig(s.cfg.Safety)
	if err := safety.ValidateMission(state.Home, waypoints, limits); err != nil {
		return err
	}
	return s.client.UploadMission(ctx, waypoints)
}

// ClearMission removes stored mission items when disarmed.
func (s *Service) ClearMission(ctx context.Context) error {
	if s.State().Armed {
		return fmt.Errorf("disarm the vehicle before clearing a mission")
	}
	return s.client.ClearMission(ctx)
}
