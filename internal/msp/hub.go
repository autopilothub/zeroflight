package msp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
)

// HubConfig holds MSP link settings.
type HubConfig struct {
	Device string
	Baud   int
	PollHz int
}

// Hub polls INAV over MSP and implements link.FC.
type Hub struct {
	client   *Client
	interval time.Duration

	mu          sync.RWMutex
	state       inav.VehicleState
	ready       bool
	boxesLoaded bool
	armBox      int
	gcsNavBox   int
}

// NewHub connects to the FC and starts background polling.
func NewHub(ctx context.Context, cfg HubConfig) (*Hub, error) {
	baud := cfg.Baud
	if baud == 0 {
		baud = 115200
	}
	pollHz := cfg.PollHz
	if pollHz <= 0 {
		pollHz = 10
	}

	client, err := Open(Config{Device: cfg.Device, Baud: baud})
	if err != nil {
		return nil, err
	}

	h := &Hub{
		client:   client,
		interval: time.Second / time.Duration(pollHz),
	}
	h.state.LinkOpen = true
	go h.run(ctx)
	return h, nil
}

func (h *Hub) run(ctx context.Context) {
	defer h.client.Close()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}
		if err := h.poll(ctx); err == nil {
			h.mu.Lock()
			h.ready = true
			h.state.Connected = true
			h.state.LinkOpen = true
			h.mu.Unlock()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (h *Hub) loadBoxes(ctx context.Context) {
	h.armBox = 0
	h.gcsNavBox = -1
	payload, err := h.client.Request(ctx, CmdBOXIDS, nil)
	if err != nil {
		h.boxesLoaded = true
		return
	}
	for i, id := range payload {
		switch int(id) {
		case BoxArm:
			h.armBox = i
		case BoxGCSNav:
			h.gcsNavBox = i
		}
	}
	h.boxesLoaded = true
}

func (h *Hub) poll(ctx context.Context) error {
	now := time.Now()

	statusPayload, err := h.client.Request(ctx, CmdSTATUS, nil)
	if err != nil {
		return err
	}
	status, err := ParseStatus(statusPayload)
	if err != nil {
		return err
	}

	gpsPayload, err := h.client.Request(ctx, CmdRAWGPS, nil)
	if err != nil {
		return err
	}
	gps, err := ParseGPS(gpsPayload)
	if err != nil {
		return err
	}

	attPayload, err := h.client.Request(ctx, CmdATTITUDE, nil)
	if err != nil {
		return err
	}
	att, err := ParseAttitude(attPayload)
	if err != nil {
		return err
	}

	altPayload, err := h.client.Request(ctx, CmdALTITUDE, nil)
	if err != nil {
		return err
	}
	alt, err := ParseAltitude(altPayload)
	if err != nil {
		return err
	}

	analogPayload, err := h.client.Request(ctx, CmdANALOG, nil)
	if err != nil {
		return err
	}
	analog, err := ParseAnalog(analogPayload)
	if err != nil {
		return err
	}

	imuPayload, err := h.client.Request(ctx, CmdRAWIMU, nil)
	var rawIMU inav.RawIMU
	if err == nil {
		if imu, parseErr := ParseRawIMU(imuPayload); parseErr == nil {
			rawIMU = inav.RawIMU{
				Accel: imu.Accel, Gyro: imu.Gyro, Mag: imu.Mag,
				Available: true, Time: now,
			}
		}
	}

	if !h.boxesLoaded {
		h.loadBoxes(ctx)
	}

	boxesPayload, err := h.client.Request(ctx, CmdACTIVEBOXES, nil)
	armed := status.Flags&(1<<2) != 0
	gcsNav := false
	if err == nil && len(boxesPayload) >= 4 {
		mask := uint32(boxesPayload[0]) | uint32(boxesPayload[1])<<8 |
			uint32(boxesPayload[2])<<16 | uint32(boxesPayload[3])<<24
		if h.armBox >= 0 {
			armed = mask&(1<<uint(h.armBox)) != 0
		}
		if h.gcsNavBox >= 0 {
			gcsNav = mask&(1<<uint(h.gcsNavBox)) != 0
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.state.Time = now
	h.state.Armed = armed
	h.state.GCSNavActive = gcsNav
	if gcsNav {
		h.state.Mode = inav.ModeGCSNav
	} else if armed {
		h.state.Mode = inav.ModePosHold
	} else {
		h.state.Mode = inav.ModeUnknown
	}

	deg := float32(0.0174532925)
	h.state.Attitude = inav.Attitude{
		Roll:  att.RollDeg * deg,
		Pitch: att.PitchDeg * deg,
		Yaw:   att.YawDeg * deg,
		Time:  now,
	}
	fixType := gps.FixType
	if fixType == 2 {
		fixType = 3 // INAV 3D fix → match MAVLink-style preflight checks
	}
	h.state.GPS = inav.GPSFix{
		Lat: gps.Lat, Lon: gps.Lon,
		AltM: gps.AltM, RelAltM: alt.EstAltM,
		FixType: fixType, Satellites: gps.Satellites,
		GroundSpeed: gps.SpeedMS, ClimbRate: alt.VarioMS,
		Time: now,
	}
	h.state.Battery = inav.Battery{
		VoltageV: analog.VoltageV, CurrentA: analog.CurrentA, Time: now,
	}
	h.state.Sensors = inav.SensorHealth{
		Gyro:  true,
		Accel: status.SensorPresent&(1<<0) != 0,
		Mag:   status.SensorPresent&(1<<2) != 0,
		Baro:  status.SensorPresent&(1<<1) != 0,
		GPS:   status.SensorPresent&(1<<3) != 0,
		Time:  now,
	}
	h.state.RawIMU = rawIMU
	return nil
}

// State returns the latest vehicle snapshot.
func (h *Hub) State() inav.VehicleState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

// Close stops polling and releases the serial port.
func (h *Hub) Close() error {
	return h.client.Close()
}

// WaitReady blocks until MSP telemetry is available.
func (h *Hub) WaitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		h.mu.RLock()
		ready := h.ready
		h.mu.RUnlock()
		if ready {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return fmt.Errorf("timed out waiting for INAV MSP telemetry")
}

// SendGoto uploads waypoint #255 via MSP_SET_WP (requires GCS NAV on RC).
func (h *Hub) SendGoto(req inav.GotoRequest) error {
	state := h.State()
	if !state.Connected {
		return fmt.Errorf("not connected to flight controller")
	}
	payload := EncodeSetWP(255, req.LatDeg, req.LonDeg, req.AltM, 0)
	_, err := h.client.Request(context.Background(), CmdSETWP, payload)
	if err != nil {
		return fmt.Errorf("msp set wp: %w", err)
	}
	return nil
}

// UploadMission uploads waypoints with MSP_SET_WP.
func (h *Hub) UploadMission(ctx context.Context, waypoints []mission.Waypoint) error {
	if h.State().Armed {
		return fmt.Errorf("cannot upload mission while armed")
	}
	for i, wp := range waypoints {
		payload := EncodeSetWP(uint8(i+1), wp.Lat, wp.Lon, wp.Alt, 0)
		if _, err := h.client.Request(ctx, CmdSETWP, payload); err != nil {
			return fmt.Errorf("msp set wp %d: %w", i+1, err)
		}
	}
	return nil
}

// ClearMission clears stored waypoints by overwriting WP1 with zeros.
func (h *Hub) ClearMission(ctx context.Context) error {
	if h.State().Armed {
		return fmt.Errorf("cannot clear mission while armed")
	}
	payload := EncodeSetWP(1, 0, 0, 0, 0)
	_, err := h.client.Request(ctx, CmdSETWP, payload)
	return err
}
