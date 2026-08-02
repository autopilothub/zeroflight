package msp

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"go.bug.st/serial"
)

// Config holds MSP serial settings.
type Config struct {
	Device string
	Baud   int
}

// Client reads MSP responses from a serial port.
type Client struct {
	port serial.Port
	mu   sync.Mutex
}

// Open connects to the MSP serial device.
func Open(cfg Config) (*Client, error) {
	baud := cfg.Baud
	if baud == 0 {
		baud = 115200
	}
	if cfg.Device == "" {
		return nil, fmt.Errorf("msp device is required")
	}
	port, err := serial.Open(cfg.Device, &serial.Mode{BaudRate: baud})
	if err != nil {
		return nil, fmt.Errorf("open msp serial: %w", err)
	}
	return &Client{port: port}, nil
}

// Close releases the serial port.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.port == nil {
		return nil
	}
	err := c.port.Close()
	c.port = nil
	return err
}

// Request sends an MSP command and returns the response payload.
func (c *Client) Request(ctx context.Context, cmd uint8, payload []byte) ([]byte, error) {
	frame, err := c.transact(ctx, cmd, EncodeRequest(cmd, payload))
	if err != nil {
		return nil, err
	}
	gotCmd, resp, err := ParseResponse(frame)
	if err != nil {
		return nil, err
	}
	if gotCmd != cmd {
		return nil, fmt.Errorf("unexpected msp command %d (expected %d)", gotCmd, cmd)
	}
	return resp, nil
}

// RequestRawIMU sends MSP_RAW_IMU and returns parsed sensor data.
func (c *Client) RequestRawIMU(ctx context.Context) (RawIMU, error) {
	payload, err := c.Request(ctx, CmdRAWIMU, nil)
	if err != nil {
		return RawIMU{}, err
	}
	return ParseRawIMU(payload)
}

func (c *Client) transact(ctx context.Context, expectCmd uint8, request []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.port == nil {
		return nil, fmt.Errorf("msp port closed")
	}

	if err := c.port.ResetInputBuffer(); err != nil {
		return nil, err
	}
	if _, err := c.port.Write(request); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	buf := make([]byte, 0, 128)
	tmp := make([]byte, 128)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := c.port.Read(tmp)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if frame, ok := extractFrame(buf, expectCmd); ok {
				return frame, nil
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil, fmt.Errorf("msp response timeout for cmd %d", expectCmd)
}

func extractFrame(buf []byte, expectCmd uint8) ([]byte, bool) {
	for i := 0; i+6 <= len(buf); i++ {
		if buf[i] != header0 || buf[i+1] != header1 || buf[i+2] != dirFrom {
			continue
		}
		size := int(buf[i+3])
		end := i + 6 + size
		if end > len(buf) {
			continue
		}
		frame := buf[i:end]
		if frame[4] != expectCmd {
			continue
		}
		if _, _, err := ParseResponse(frame); err == nil {
			return frame, true
		}
	}
	return nil, false
}
