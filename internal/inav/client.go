package inav

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

// Config holds MAVLink connection settings for INAV.
type Config struct {
	Device           string
	Baud             int
	UDPAddress       string
	TargetSystemID   uint8
	TargetComponentID uint8
	OutSystemID      uint8
	OutComponentID   uint8
	MAVLinkVersion   gomavlib.Version
}

// DefaultConfig returns settings suited for Mamba F405 MK2 + INAV on UART6.
func DefaultConfig() Config {
	return Config{
		Device:            "/dev/serial0",
		Baud:              115200,
		TargetSystemID:    1,
		TargetComponentID: 1,
		OutSystemID:       255,
		OutComponentID:    190,
		MAVLinkVersion:    gomavlib.V2,
	}
}

// Client communicates with INAV over MAVLink.
type Client struct {
	cfg Config

	mu          sync.RWMutex
	state       VehicleState
	channel     *gomavlib.Channel
	missionXfer *missionTransfer

	node *gomavlib.Node
}

// NewClient creates an INAV MAVLink client.
func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg}
}

// Connect opens the MAVLink link and starts the event loop.
func (c *Client) Connect(ctx context.Context) error {
	endpoints, err := c.buildEndpoints()
	if err != nil {
		return err
	}

	node := &gomavlib.Node{
		Endpoints:       endpoints,
		Dialect:         ardupilotmega.Dialect,
		OutVersion:      c.cfg.MAVLinkVersion,
		OutSystemID:     c.cfg.OutSystemID,
		OutComponentID:  c.cfg.OutComponentID,
	}
	if err := node.Initialize(); err != nil {
		return fmt.Errorf("initialize mavlink node: %w", err)
	}

	c.node = node
	go c.run(ctx)
	return nil
}

func (c *Client) buildEndpoints() ([]gomavlib.EndpointConf, error) {
	if c.cfg.UDPAddress != "" {
		return []gomavlib.EndpointConf{
			gomavlib.EndpointUDPClient{Address: c.cfg.UDPAddress},
		}, nil
	}
	if c.cfg.Device == "" {
		return nil, fmt.Errorf("serial device is required when udp address is empty")
	}
	baud := c.cfg.Baud
	if baud == 0 {
		baud = 115200
	}
	return []gomavlib.EndpointConf{
		gomavlib.EndpointSerial{
			Device: c.cfg.Device,
			Baud:   baud,
		},
	}, nil
}

func (c *Client) run(ctx context.Context) {
	defer func() {
		if c.node != nil {
			c.node.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-c.node.Events():
			if !ok {
				return
			}
			c.handleEvent(evt)
		}
	}
}

func (c *Client) handleEvent(evt gomavlib.Event) {
	switch evt := evt.(type) {
	case *gomavlib.EventFrame:
		c.mu.Lock()
		c.channel = evt.Channel
		channel := evt.Channel
		msg := evt.Message()
		c.applyFrame(msg)
		c.mu.Unlock()
		c.handleMissionFrame(channel, msg)

	case *gomavlib.EventChannelOpen:
		c.mu.Lock()
		c.state.Connected = true
		c.state.Time = time.Now()
		c.mu.Unlock()

	case *gomavlib.EventChannelClose:
		c.mu.Lock()
		c.state.Connected = false
		c.state.Time = time.Now()
		c.mu.Unlock()
	}
}

func (c *Client) applyFrame(msg any) {
	switch m := msg.(type) {
	case *ardupilotmega.MessageHeartbeat:
		applyHeartbeat(&c.state, m)
	case *ardupilotmega.MessageAttitude:
		applyAttitude(&c.state, m)
	case *ardupilotmega.MessageGpsRawInt:
		applyGPSRaw(&c.state, m)
	case *ardupilotmega.MessageGlobalPositionInt:
		applyGlobalPosition(&c.state, m)
	case *ardupilotmega.MessageSysStatus:
		applySysStatus(&c.state, m)
	case *ardupilotmega.MessageGpsGlobalOrigin:
		applyGpsGlobalOrigin(&c.state, m)
	}
}

// State returns a copy of the latest vehicle state.
func (c *Client) State() VehicleState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Close shuts down the MAVLink node.
func (c *Client) Close() {
	if c.node != nil {
		c.node.Close()
	}
}

// WaitForConnection blocks until telemetry is received or the context is canceled.
func (c *Client) WaitForConnection(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state := c.State()
		if state.Connected && !state.Time.IsZero() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return fmt.Errorf("timed out waiting for INAV telemetry")
}

// ParseConnection parses "serial:/dev/serial0:115200" or "udp:127.0.0.1:14550".
func ParseConnection(raw string, base Config) (Config, error) {
	cfg := base
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		return cfg, fmt.Errorf("invalid connection %q: use serial:device:baud or udp:host:port", raw)
	}

	switch parts[0] {
	case "serial":
		cfg.UDPAddress = ""
		cfg.Device = parts[1]
		if len(parts) == 3 {
			var baud int
			if _, err := fmt.Sscanf(parts[2], "%d", &baud); err != nil {
				return cfg, fmt.Errorf("invalid baud %q: %w", parts[2], err)
			}
			cfg.Baud = baud
		}
	case "udp":
		cfg.Device = ""
		if len(parts) == 3 {
			cfg.UDPAddress = fmt.Sprintf("%s:%s", parts[1], parts[2])
		} else {
			return cfg, fmt.Errorf("invalid udp connection %q: use udp:host:port", raw)
		}
	default:
		return cfg, fmt.Errorf("unsupported connection scheme %q", parts[0])
	}
	return cfg, nil
}
