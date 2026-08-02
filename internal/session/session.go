package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/autopilothub/zeroflight/internal/config"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/link"
	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/msp"
)

// Session holds the active flight-controller link (MSP or MAVLink).
type Session struct {
	cfg    config.File
	fc     link.FC
	cancel context.CancelFunc

	closeOnce sync.Once
}

// Open connects to the flight controller using the configured protocol.
func Open(ctx context.Context, cfgPath, connectionOverride string) (*Session, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctx)
	fc, err := openFC(runCtx, cfg, connectionOverride)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Session{cfg: cfg, fc: fc, cancel: cancel}, nil
}

func openFC(ctx context.Context, cfg config.File, connectionOverride string) (link.FC, error) {
	linkCfg := cfg.LinkConfig(connectionOverride)
	switch linkCfg.Protocol {
	case "mavlink":
		clientCfg, err := cfg.INAVConfig(connectionOverride)
		if err != nil {
			return nil, err
		}
		client := inav.NewClient(clientCfg)
		if err := client.Connect(ctx); err != nil {
			return nil, fmt.Errorf("connect mavlink: %w", err)
		}
		return &inav.MAVLinkAdapter{Client: client}, nil
	default:
		hub, err := msp.NewHub(ctx, msp.HubConfig{
			Device: linkCfg.Device,
			Baud:   linkCfg.Baud,
			PollHz: linkCfg.PollHz,
		})
		if err != nil {
			return nil, fmt.Errorf("connect msp: %w", err)
		}
		return hub, nil
	}
}

// Close shuts down the flight-controller link.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.fc != nil {
			_ = s.fc.Close()
		}
	})
}

// Config returns loaded configuration.
func (s *Session) Config() config.File {
	return s.cfg
}

// State returns the latest vehicle snapshot.
func (s *Session) State() inav.VehicleState {
	return s.fc.State()
}

// WaitReady blocks until telemetry is available.
func (s *Session) WaitReady(ctx context.Context) error {
	return s.fc.WaitReady(ctx, 15*time.Second)
}

// SendGoto sends a navigation command to the FC.
func (s *Session) SendGoto(req inav.GotoRequest) error {
	return s.fc.SendGoto(req)
}

// UploadMission uploads waypoints to the FC.
func (s *Session) UploadMission(ctx context.Context, waypoints []mission.Waypoint) error {
	return s.fc.UploadMission(ctx, waypoints)
}

// ClearMission clears stored mission waypoints.
func (s *Session) ClearMission(ctx context.Context) error {
	return s.fc.ClearMission(ctx)
}
