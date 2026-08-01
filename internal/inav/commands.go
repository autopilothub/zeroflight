package inav

import (
	"fmt"
	"math"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

// SendGoto sends MAV_CMD_DO_REPOSITION to INAV waypoint #255.
// INAV requires POS HOLD + GCS NAV to be active before accepting this command.
func (c *Client) SendGoto(req GotoRequest) error {
	c.mu.RLock()
	node := c.node
	channel := c.channel
	c.mu.RUnlock()

	if node == nil {
		return fmt.Errorf("mavlink node is not initialized")
	}
	if channel == nil {
		return fmt.Errorf("no mavlink channel yet; wait for telemetry")
	}

	msg := &ardupilotmega.MessageCommandInt{
		TargetSystem:    c.cfg.TargetSystemID,
		TargetComponent: c.cfg.TargetComponentID,
		Frame:           ardupilotmega.MAV_FRAME_GLOBAL,
		Command:         common.MAV_CMD(ardupilotmega.MAV_CMD_DO_REPOSITION),
		X:               int32(math.Round(req.LatDeg * 1e7)),
		Y:               int32(math.Round(req.LonDeg * 1e7)),
		Z:               req.AltM,
	}
	if req.YawDeg != nil {
		msg.Param4 = *req.YawDeg
	}

	if err := node.WriteMessageTo(channel, msg); err != nil {
		return fmt.Errorf("send DO_REPOSITION: %w", err)
	}
	return nil
}

// PreflightGoto checks whether INAV is ready to accept a goto command.
func PreflightGoto(state VehicleState) error {
	if !state.Connected {
		return fmt.Errorf("not connected to flight controller")
	}
	if !state.Armed {
		return fmt.Errorf("vehicle is not armed")
	}
	if state.GPS.FixType < 3 {
		return fmt.Errorf("GPS 3D fix required (current fix type: %d)", state.GPS.FixType)
	}
	if state.GPS.Satellites < 6 {
		return fmt.Errorf("insufficient satellites: %d (need >= 6)", state.GPS.Satellites)
	}
	if !state.GCSNavActive {
		return fmt.Errorf("GCS NAV mode is not active; enable GCS NAV on the transmitter before goto")
	}
	return nil
}
