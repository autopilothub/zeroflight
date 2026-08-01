package inav

import "github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"

// ParseCopterMode maps ArduPilot-style custom_mode from INAV HEARTBEAT to FlightMode.
func ParseCopterMode(customMode uint32) (FlightMode, bool) {
	switch ardupilotmega.COPTER_MODE(customMode) {
	case ardupilotmega.COPTER_MODE_STABILIZE:
		return ModeStabilize, true
	case ardupilotmega.COPTER_MODE_ACRO:
		return ModeAcro, true
	case ardupilotmega.COPTER_MODE_ALT_HOLD:
		return ModeAltHold, true
	case ardupilotmega.COPTER_MODE_GUIDED:
		return ModeGCSNav, true
	case ardupilotmega.COPTER_MODE_LOITER:
		return ModePosHold, true
	case ardupilotmega.COPTER_MODE_AUTO:
		return ModeMission, true
	case ardupilotmega.COPTER_MODE_RTL:
		return ModeRTH, true
	case ardupilotmega.COPTER_MODE_LAND:
		return ModeLand, true
	default:
		return ModeUnknown, false
	}
}
