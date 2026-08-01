package session

import (
	"context"
	"fmt"
	"time"

	"github.com/autopilothub/zeroflight/internal/config"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/msp"
)

// Session holds MAVLink and optional MSP connections.
type Session struct {
	cfg    config.File
	client *inav.Client
	msp    *msp.Poller
	cancel context.CancelFunc
}

// Open connects MAVLink and optionally starts an MSP poller.
func Open(ctx context.Context, cfgPath, connectionOverride string) (*Session, error) {
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
		return nil, fmt.Errorf("connect mavlink: %w", err)
	}

	s := &Session{cfg: cfg, client: client, cancel: cancel}

	if cfg.MSP.Enabled {
		mspClient, err := msp.Open(msp.Config{Device: cfg.MSP.Device, Baud: cfg.MSP.Baud})
		if err != nil {
			s.Close()
			return nil, err
		}
		interval := time.Second / time.Duration(cfg.MSP.PollHz)
		if cfg.MSP.PollHz <= 0 {
			interval = 100 * time.Millisecond
		}
		poller := msp.NewPoller(mspClient, interval)
		go func() {
			poller.Run(runCtx)
			_ = mspClient.Close()
		}()
		s.msp = poller
	}

	return s, nil
}

// Close shuts down background workers and connections.
func (s *Session) Close() {
	s.cancel()
	s.client.Close()
}

// Config returns loaded configuration.
func (s *Session) Config() config.File {
	return s.cfg
}

// Client returns the MAVLink client.
func (s *Session) Client() *inav.Client {
	return s.client
}

// State returns merged MAVLink and MSP telemetry.
func (s *Session) State() inav.VehicleState {
	state := s.client.State()
	if s.msp != nil {
		state.RawIMU = s.msp.Latest()
	}
	return state
}

// WaitReady blocks until MAVLink telemetry is available.
func (s *Session) WaitReady(ctx context.Context) error {
	return s.client.WaitForConnection(ctx, 15*time.Second)
}
