package msp_test

import (
	"testing"

	"github.com/autopilothub/zeroflight/internal/msp"
)

func TestParseGPS(t *testing.T) {
	payload := []byte{
		2, 10, // fix, sats
		0x00, 0x96, 0x14, 0x06, // lat
		0x00, 0x28, 0x6E, 0x1F, // lon
		0x64, 0x00, // alt 100m
		0x64, 0x00, // speed 100 cm/s = 1 m/s
		0xB4, 0x00, // course 18.0 deg
	}
	gps, err := msp.ParseGPS(payload)
	if err != nil {
		t.Fatal(err)
	}
	if gps.FixType != 2 || gps.Satellites != 10 {
		t.Fatalf("unexpected fix/sats: %+v", gps)
	}
	if gps.SpeedMS < 0.99 || gps.SpeedMS > 1.01 {
		t.Fatalf("unexpected speed: %f", gps.SpeedMS)
	}
}

func TestParseStatusSensors(t *testing.T) {
	payload := make([]byte, 11)
	payload[4] = 0x0F // accel+baro+mag+gps bits
	payload[5] = 0x00
	payload[6] = 0x04 // armed flag bit 2
	status, err := msp.ParseStatus(payload)
	if err != nil {
		t.Fatal(err)
	}
	if status.SensorPresent&0x0F != 0x0F {
		t.Fatalf("sensor present: %x", status.SensorPresent)
	}
	if status.Flags&(1<<2) == 0 {
		t.Fatal("expected armed flag")
	}
}

func TestEncodeSetWP(t *testing.T) {
	payload := msp.EncodeSetWP(255, 37.5665, 126.9780, 15, 0)
	if len(payload) != 21 || payload[0] != 255 {
		t.Fatalf("unexpected payload len=%d wp=%d", len(payload), payload[0])
	}
}
