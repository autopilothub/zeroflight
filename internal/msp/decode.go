package msp

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Status holds MSP_STATUS fields.
type Status struct {
	SensorPresent uint16
	Flags         uint32
}

// ParseStatus decodes MSP_STATUS (minimum 11 bytes).
func ParseStatus(payload []byte) (Status, error) {
	if len(payload) < 11 {
		return Status{}, fmt.Errorf("msp status payload too short (%d bytes)", len(payload))
	}
	return Status{
		SensorPresent: binary.LittleEndian.Uint16(payload[4:6]),
		Flags:         binary.LittleEndian.Uint32(payload[6:10]),
	}, nil
}

// GPS holds MSP_RAW_GPS fields.
type GPS struct {
	FixType    uint8
	Satellites uint8
	Lat        float64
	Lon        float64
	AltM       float32
	SpeedMS    float32
	CourseDeg  float32
}

// ParseGPS decodes MSP_RAW_GPS.
func ParseGPS(payload []byte) (GPS, error) {
	if len(payload) < 16 {
		return GPS{}, fmt.Errorf("msp raw gps payload too short (%d bytes)", len(payload))
	}
	lat := float64(int32(binary.LittleEndian.Uint32(payload[2:6]))) / 1e7
	lon := float64(int32(binary.LittleEndian.Uint32(payload[6:10]))) / 1e7
	alt := float32(int16(binary.LittleEndian.Uint16(payload[10:12])))
	speed := float32(binary.LittleEndian.Uint16(payload[12:14])) / 100
	course := float32(binary.LittleEndian.Uint16(payload[14:16])) / 10
	return GPS{
		FixType:    payload[0],
		Satellites: payload[1],
		Lat:        lat,
		Lon:        lon,
		AltM:       alt,
		SpeedMS:    speed,
		CourseDeg:  course,
	}, nil
}

// Attitude holds MSP_ATTITUDE fields (0.1 degree units).
type Attitude struct {
	RollDeg  float32
	PitchDeg float32
	YawDeg   float32
}

// ParseAttitude decodes MSP_ATTITUDE.
func ParseAttitude(payload []byte) (Attitude, error) {
	if len(payload) < 6 {
		return Attitude{}, fmt.Errorf("msp attitude payload too short (%d bytes)", len(payload))
	}
	scale := float32(0.1)
	return Attitude{
		RollDeg:  float32(int16(binary.LittleEndian.Uint16(payload[0:2]))) * scale,
		PitchDeg: float32(int16(binary.LittleEndian.Uint16(payload[2:4]))) * scale,
		YawDeg:   float32(int16(binary.LittleEndian.Uint16(payload[4:6]))) * scale,
	}, nil
}

// Altitude holds MSP_ALTITUDE fields.
type Altitude struct {
	EstAltM   float32
	VarioMS   float32
}

// ParseAltitude decodes MSP_ALTITUDE.
func ParseAltitude(payload []byte) (Altitude, error) {
	if len(payload) < 6 {
		return Altitude{}, fmt.Errorf("msp altitude payload too short (%d bytes)", len(payload))
	}
	return Altitude{
		EstAltM: float32(int32(binary.LittleEndian.Uint32(payload[0:4]))) / 100,
		VarioMS: float32(int16(binary.LittleEndian.Uint16(payload[4:6]))) / 100,
	}, nil
}

// Analog holds MSP_ANALOG battery fields.
type Analog struct {
	VoltageV float32
	CurrentA float32
}

// ParseAnalog decodes MSP_ANALOG.
func ParseAnalog(payload []byte) (Analog, error) {
	if len(payload) < 7 {
		return Analog{}, fmt.Errorf("msp analog payload too short (%d bytes)", len(payload))
	}
	vbat := float32(payload[0]) / 10
	amps := float32(binary.LittleEndian.Uint16(payload[5:7])) / 100
	return Analog{VoltageV: vbat, CurrentA: amps}, nil
}

// EncodeSetWP builds MSP_SET_WP payload for INAV.
func EncodeSetWP(wpNum uint8, lat, lon float64, altM float32, flags uint16) []byte {
	buf := make([]byte, 21)
	buf[0] = wpNum
	buf[1] = 0x01 // NAV_WP_ACTION_WAYPOINT
	binary.LittleEndian.PutUint32(buf[2:6], uint32(math.Round(lat*1e7)))
	binary.LittleEndian.PutUint32(buf[6:10], uint32(math.Round(lon*1e7)))
	binary.LittleEndian.PutUint32(buf[10:14], uint32(math.Round(float64(altM)*100)))
	binary.LittleEndian.PutUint16(buf[18:20], flags)
	return buf
}
