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

// MAVLinkVersion converts config value (1 or 2) to gomavlib version.
func MAVLinkVersion(v int) gomavlib.Version {
	if v == 1 {
		return gomavlib.V1
	}
	return gomavlib.V2
}

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
	streamsOnce sync.Once

	node      *gomavlib.Node
	closeOnce sync.Once
	done      chan struct{}
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
		Endpoints:              endpoints,
		Dialect:                ardupilotmega.Dialect,
		OutVersion:             c.cfg.MAVLinkVersion,
		OutSystemID:            c.cfg.OutSystemID,
		OutComponentID:         c.cfg.OutComponentID,
		HeartbeatPeriod:        time.Second,
		StreamRequestEnable:    true,
		StreamRequestFrequency: 10,
	}
	if err := node.Initialize(); err != nil {
		return fmt.Errorf("initialize mavlink node: %w", err)
	}

	c.node = node
	c.done = make(chan struct{})
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
	defer close(c.done)
	defer c.closeNode()

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

func (c *Client) closeNode() {
	c.closeOnce.Do(func() {
		if c.node != nil {
			c.node.Close()
			c.node = nil
		}
	})
}

func (c *Client) handleEvent(evt gomavlib.Event) {
	switch evt := evt.(type) {
	case *gomavlib.EventFrame:
		c.mu.Lock()
		c.channel = evt.Channel
		channel := evt.Channel
		msg := evt.Message()
		if hb, ok := msg.(*ardupilotmega.MessageHeartbeat); ok {
			applyHeartbeat(&c.state, hb)
			c.mu.Unlock()
			c.onHeartbeat(channel)
		} else {
			c.applyFrame(msg)
			c.mu.Unlock()
		}
		c.handleMissionFrame(channel, msg)

	case *gomavlib.EventChannelOpen:
		c.mu.Lock()
		c.channel = evt.Channel
		c.state.LinkOpen = true
		c.state.Time = time.Now()
		c.mu.Unlock()

	case *gomavlib.EventChannelClose:
		c.mu.Lock()
		c.state.LinkOpen = false
		c.state.Connected = false
		c.state.Time = time.Now()
		c.mu.Unlock()

	case *gomavlib.EventParseError:
		c.mu.Lock()
		c.state.ParseErrors++
		c.state.LinkOpen = true
		if evt.Channel != nil {
			c.channel = evt.Channel
		}
		c.state.Time = time.Now()
		c.mu.Unlock()
	}
}

func (c *Client) onHeartbeat(channel *gomavlib.Channel) {
	c.streamsOnce.Do(func() {
		c.requestTelemetryStreams(channel)
	})
}

func (c *Client) applyFrame(msg any) {
	switch m := msg.(type) {
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

// Close waits for the event loop to exit and shuts down the MAVLink node once.
func (c *Client) Close() {
	if c.done != nil {
		<-c.done
	}
	c.closeNode()
}

// WaitForConnection blocks until an FC HEARTBEAT is received or the context is canceled.
func (c *Client) WaitForConnection(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state := c.State()
		if state.Connected {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	state := c.State()
	if state.ParseErrors > 0 {
		return fmt.Errorf("timed out waiting for INAV heartbeat (%d mavlink parse errors; check baud and mavlink_version)", state.ParseErrors)
	}
	if state.LinkOpen {
		return fmt.Errorf("timed out waiting for INAV heartbeat (serial open but no FC data; check UART6 MAVLink wiring and INAV ports)")
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
