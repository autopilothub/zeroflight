package inav

import "time"

// FlightMode is the interpreted INAV flight mode from MAVLink HEARTBEAT.
type FlightMode string

const (
	ModeUnknown   FlightMode = "UNKNOWN"
	ModeManual    FlightMode = "MANUAL"
	ModeAcro      FlightMode = "ACRO"
	ModeStabilize FlightMode = "STABILIZE"
	ModeAltHold   FlightMode = "ALT_HOLD"
	ModePosHold   FlightMode = "POS_HOLD"
	ModeGCSNav    FlightMode = "GCS_NAV"
	ModeMission   FlightMode = "MISSION"
	ModeRTH       FlightMode = "RTH"
	ModeLand      FlightMode = "LAND"
	ModeFailsafe  FlightMode = "FAILSAFE"
)

// Attitude holds roll/pitch/yaw and body rates from MAVLink ATTITUDE.
type Attitude struct {
	Roll       float32
	Pitch      float32
	Yaw        float32
	RollSpeed  float32
	PitchSpeed float32
	YawSpeed   float32
	Time       time.Time
}

// GPSFix holds GNSS data from GPS_RAW_INT and GLOBAL_POSITION_INT.
type GPSFix struct {
	Lat         float64
	Lon         float64
	AltM        float32
	RelAltM     float32
	FixType     uint8
	Satellites  uint8
	HDOP        float32
	GroundSpeed float32
	ClimbRate   float32
	Time        time.Time
}

// Battery holds power state from SYS_STATUS / BATTERY_STATUS.
type Battery struct {
	VoltageV    float32
	CurrentA    float32
	RemainingPct int8
	Time        time.Time
}

// SensorHealth reports which onboard sensors INAV considers present.
type SensorHealth struct {
	Gyro     bool
	Accel    bool
	Mag      bool
	Baro     bool
	GPS      bool
	RC       bool
	Time     time.Time
}

// HomePosition is the INAV home / arming point from GPS_GLOBAL_ORIGIN.
type HomePosition struct {
	Lat   float64
	Lon   float64
	AltM  float32
	Valid bool
	Time  time.Time
}

// RawIMU holds raw accelerometer, gyro, and magnetometer samples (typically from MSP).
type RawIMU struct {
	Accel     [3]int16
	Gyro      [3]int16
	Mag       [3]int16
	Available bool
	Time      time.Time
}

// VehicleState is the latest aggregated telemetry snapshot.
type VehicleState struct {
	Time         time.Time
	LinkOpen     bool // serial/UDP channel is open
	Connected    bool // FC HEARTBEAT received (real MAVLink link)
	ParseErrors  uint64
	Armed        bool
	Mode         FlightMode
	GCSNavActive bool
	Attitude     Attitude
	GPS          GPSFix
	Battery      Battery
	Sensors      SensorHealth
	Home         HomePosition
	RawIMU       RawIMU
}

// GotoRequest is a target position for MAV_CMD_DO_REPOSITION.
type GotoRequest struct {
	LatDeg  float64
	LonDeg  float64
	AltM    float32
	YawDeg  *float32
}
