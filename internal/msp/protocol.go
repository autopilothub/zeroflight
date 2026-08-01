package msp

import "fmt"

const (
	header0 = '$'
	header1 = 'M'
	dirTo   = '<'
	dirFrom = '>'
)

// MSP command IDs used by ZeroFlight.
const (
	CmdRAWIMU = 102
)

// EncodeRequest builds an MSP v1 request frame.
func EncodeRequest(cmd uint8, payload []byte) []byte {
	size := uint8(len(payload))
	frame := make([]byte, 6+len(payload))
	frame[0] = header0
	frame[1] = header1
	frame[2] = dirTo
	frame[3] = size
	frame[4] = cmd
	copy(frame[5:], payload)
	frame[len(frame)-1] = checksum(size, cmd, payload)
	return frame
}

func checksum(size, cmd uint8, payload []byte) uint8 {
	crc := size ^ cmd
	for _, b := range payload {
		crc ^= b
	}
	return crc
}

// ParseResponse validates and extracts the payload from an MSP v1 response frame.
func ParseResponse(frame []byte) (cmd uint8, payload []byte, err error) {
	if len(frame) < 6 {
		return 0, nil, fmt.Errorf("msp frame too short (%d bytes)", len(frame))
	}
	if frame[0] != header0 || frame[1] != header1 || frame[2] != dirFrom {
		return 0, nil, fmt.Errorf("invalid msp response header")
	}
	size := frame[3]
	cmd = frame[4]
	if int(size)+6 != len(frame) {
		return 0, nil, fmt.Errorf("msp size mismatch: declared %d frame %d", size, len(frame))
	}
	payload = frame[5 : 5+size]
	if checksum(size, cmd, payload) != frame[len(frame)-1] {
		return 0, nil, fmt.Errorf("msp checksum mismatch")
	}
	return cmd, payload, nil
}

// RawIMU holds accelerometer, gyro, and magnetometer raw values from MSP_RAW_IMU.
type RawIMU struct {
	Accel [3]int16
	Gyro  [3]int16
	Mag   [3]int16
}

// ParseRawIMU decodes the 18-byte MSP_RAW_IMU payload.
func ParseRawIMU(payload []byte) (RawIMU, error) {
	if len(payload) < 18 {
		return RawIMU{}, fmt.Errorf("msp raw imu payload too short (%d bytes)", len(payload))
	}
	var imu RawIMU
	for i := 0; i < 3; i++ {
		imu.Accel[i] = int16(uint16(payload[i*2]) | uint16(payload[i*2+1])<<8)
		imu.Gyro[i] = int16(uint16(payload[6+i*2]) | uint16(payload[6+i*2+1])<<8)
		imu.Mag[i] = int16(uint16(payload[12+i*2]) | uint16(payload[12+i*2+1])<<8)
	}
	return imu, nil
}
