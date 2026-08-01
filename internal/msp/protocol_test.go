package msp_test

import (
	"testing"

	"github.com/autopilothub/zeroflight/internal/msp"
)

func TestEncodeRequestChecksum(t *testing.T) {
	frame := msp.EncodeRequest(msp.CmdRAWIMU, nil)
	if len(frame) != 6 || frame[4] != msp.CmdRAWIMU {
		t.Fatalf("unexpected frame: %v", frame)
	}
}

func TestParseRawIMU(t *testing.T) {
	payload := []byte{
		0x01, 0x00, 0x02, 0x00, 0x03, 0x00,
		0x04, 0x00, 0x05, 0x00, 0x06, 0x00,
		0x07, 0x00, 0x08, 0x00, 0x09, 0x00,
	}
	imu, err := msp.ParseRawIMU(payload)
	if err != nil {
		t.Fatal(err)
	}
	if imu.Accel[0] != 1 || imu.Gyro[2] != 6 || imu.Mag[2] != 9 {
		t.Fatalf("unexpected imu: %+v", imu)
	}
}

func TestParseResponseRoundTrip(t *testing.T) {
	req := msp.EncodeRequest(msp.CmdRAWIMU, nil)
	_ = req
	payload := make([]byte, 18)
	resp := msp.EncodeRequest(msp.CmdRAWIMU, nil)
	resp[2] = '>'
	resp = append([]byte{'$', 'M', '>'}, resp[3:]...)
	// build valid response manually
	size := byte(18)
	cmd := byte(msp.CmdRAWIMU)
	frame := []byte{'$', 'M', '>'}
	frame = append(frame, size, cmd)
	frame = append(frame, payload...)
	crc := size ^ cmd
	for _, b := range payload {
		crc ^= b
	}
	frame = append(frame, crc)

	gotCmd, gotPayload, err := msp.ParseResponse(frame)
	if err != nil {
		t.Fatal(err)
	}
	if gotCmd != msp.CmdRAWIMU || len(gotPayload) != 18 {
		t.Fatalf("unexpected response cmd=%d len=%d", gotCmd, len(gotPayload))
	}
}
