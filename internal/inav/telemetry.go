package inav

import (
	"time"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

func applyHeartbeat(state *VehicleState, msg *ardupilotmega.MessageHeartbeat) {
	now := time.Now()
	state.Time = now
	state.Connected = true
	state.Armed = msg.BaseMode&ardupilotmega.MAV_MODE_FLAG_SAFETY_ARMED != 0

	mode, ok := ParseCopterMode(msg.CustomMode)
	if ok {
		state.Mode = mode
		state.GCSNavActive = mode == ModeGCSNav
	}
}

func applyAttitude(state *VehicleState, msg *ardupilotmega.MessageAttitude) {
	now := time.Now()
	state.Attitude = Attitude{
		Roll:       msg.Roll,
		Pitch:      msg.Pitch,
		Yaw:        msg.Yaw,
		RollSpeed:  msg.Rollspeed,
		PitchSpeed: msg.Pitchspeed,
		YawSpeed:   msg.Yawspeed,
		Time:       now,
	}
	state.Time = now
}

func applyGPSRaw(state *VehicleState, msg *ardupilotmega.MessageGpsRawInt) {
	now := time.Now()
	state.GPS.Lat = float64(msg.Lat) / 1e7
	state.GPS.Lon = float64(msg.Lon) / 1e7
	state.GPS.AltM = float32(msg.Alt) / 1000
	state.GPS.FixType = uint8(msg.FixType)
	state.GPS.Satellites = msg.SatellitesVisible
	state.GPS.HDOP = float32(msg.Eph) / 100
	state.GPS.GroundSpeed = float32(msg.Vel) / 100
	state.GPS.Time = now
	state.Time = now
}

func applyGlobalPosition(state *VehicleState, msg *ardupilotmega.MessageGlobalPositionInt) {
	now := time.Now()
	if msg.Lat != 0 || msg.Lon != 0 {
		state.GPS.Lat = float64(msg.Lat) / 1e7
		state.GPS.Lon = float64(msg.Lon) / 1e7
	}
	state.GPS.AltM = float32(msg.Alt) / 1000
	state.GPS.RelAltM = float32(msg.RelativeAlt) / 1000
	state.GPS.GroundSpeed = float32(msg.Vx*msg.Vx+msg.Vy*msg.Vy) / 100
	state.GPS.ClimbRate = float32(msg.Vz) / 100
	state.GPS.Time = now
	state.Time = now
}

func applySysStatus(state *VehicleState, msg *ardupilotmega.MessageSysStatus) {
	now := time.Now()
	present := msg.OnboardControlSensorsPresent
	enabled := msg.OnboardControlSensorsEnabled

	state.Sensors.Gyro = sensorPresent(present, enabled, ardupilotmega.MAV_SYS_STATUS_SENSOR_3D_GYRO)
	state.Sensors.Accel = sensorPresent(present, enabled, ardupilotmega.MAV_SYS_STATUS_SENSOR_3D_ACCEL)
	state.Sensors.Mag = sensorPresent(present, enabled, ardupilotmega.MAV_SYS_STATUS_SENSOR_3D_MAG)
	state.Sensors.Baro = sensorPresent(present, enabled, ardupilotmega.MAV_SYS_STATUS_SENSOR_ABSOLUTE_PRESSURE)
	state.Sensors.GPS = sensorPresent(present, enabled, ardupilotmega.MAV_SYS_STATUS_SENSOR_GPS)
	state.Sensors.RC = sensorPresent(present, enabled, ardupilotmega.MAV_SYS_STATUS_SENSOR_RC_RECEIVER)
	state.Sensors.Time = now

	if msg.VoltageBattery != 65535 {
		state.Battery.VoltageV = float32(msg.VoltageBattery) / 1000
	}
	if msg.CurrentBattery != -1 {
		state.Battery.CurrentA = float32(msg.CurrentBattery) / 100
	}
	if msg.BatteryRemaining != -1 {
		state.Battery.RemainingPct = int8(msg.BatteryRemaining)
	}
	state.Battery.Time = now
	state.Time = now
}

func sensorPresent(present, enabled, flag ardupilotmega.MAV_SYS_STATUS_SENSOR) bool {
	return present&flag != 0 && enabled&flag != 0
}

func applyGpsGlobalOrigin(state *VehicleState, msg *ardupilotmega.MessageGpsGlobalOrigin) {
	now := time.Now()
	state.Home = HomePosition{
		Lat:   float64(msg.Latitude) / 1e7,
		Lon:   float64(msg.Longitude) / 1e7,
		AltM:  float32(msg.Altitude) / 1000,
		Valid: true,
		Time:  now,
	}
	state.Time = now
}
