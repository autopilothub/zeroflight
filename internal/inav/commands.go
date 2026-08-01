package inav

import (
	"fmt"
	"math"
	"strings"

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

const defaultMaxHDOP = 2.0

// GotoPreflightOptions configures goto/hover preflight checks.
type GotoPreflightOptions struct {
	Force   bool
	MaxHDOP float32
}

// GotoPreflightResult holds non-fatal warnings from preflight checks.
type GotoPreflightResult struct {
	Warnings []string
}

// CheckGotoPreflight validates vehicle state before sending goto/hover.
// Hard failures return an error. HDOP above MaxHDOP blocks unless Force is set.
func CheckGotoPreflight(state VehicleState, opts GotoPreflightOptions) (GotoPreflightResult, error) {
	maxHDOP := opts.MaxHDOP
	if maxHDOP <= 0 {
		maxHDOP = defaultMaxHDOP
	}

	if !state.Connected {
		return GotoPreflightResult{}, fmt.Errorf("not connected to flight controller")
	}
	if !state.Armed {
		return GotoPreflightResult{}, fmt.Errorf("vehicle is not armed")
	}
	if state.GPS.FixType < 3 {
		return GotoPreflightResult{}, fmt.Errorf("GPS 3D fix required (current fix type: %d)", state.GPS.FixType)
	}
	if state.GPS.Satellites < 6 {
		return GotoPreflightResult{}, fmt.Errorf("insufficient satellites: %d (need >= 6)", state.GPS.Satellites)
	}
	if !state.GCSNavActive {
		if state.Mode == ModePosHold {
			return GotoPreflightResult{}, fmt.Errorf(
				"GCS NAV is not active; enable the GCS NAV switch while in POS HOLD",
			)
		}
		return GotoPreflightResult{}, fmt.Errorf(
			"GCS NAV mode is not active (current mode: %s); enable GCS NAV on the transmitter",
			state.Mode,
		)
	}

	var result GotoPreflightResult
	if state.GPS.HDOP > maxHDOP {
		msg := fmt.Sprintf("HDOP %.1f exceeds %.1f", state.GPS.HDOP, maxHDOP)
		if !opts.Force {
			return result, fmt.Errorf("%s; use --force to proceed anyway", msg)
		}
		result.Warnings = append(result.Warnings, msg+" (forced)")
	}

	return result, nil
}

// PreflightGoto checks whether INAV is ready to accept a goto command.
func PreflightGoto(state VehicleState) error {
	_, err := CheckGotoPreflight(state, GotoPreflightOptions{})
	return err
}

// FormatWarnings joins preflight warnings for stderr output.
func FormatWarnings(warnings []string) string {
	return strings.Join(warnings, "; ")
}
