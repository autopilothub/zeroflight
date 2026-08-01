package msp

import (
	"context"
	"sync"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
)

// Poller periodically requests MSP_RAW_IMU and stores the latest sample.
type Poller struct {
	client   *Client
	interval time.Duration

	mu  sync.RWMutex
	imu inav.RawIMU
}

// NewPoller creates a background MSP poller.
func NewPoller(client *Client, interval time.Duration) *Poller {
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	return &Poller{client: client, interval: interval}
}

// Run polls until the context is canceled.
func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		imu, err := p.client.RequestRawIMU(ctx)
		if err == nil {
			p.mu.Lock()
			p.imu = inav.RawIMU{
				Accel:     imu.Accel,
				Gyro:      imu.Gyro,
				Mag:       imu.Mag,
				Available: true,
				Time:      time.Now(),
			}
			p.mu.Unlock()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Latest returns the most recent raw IMU sample.
func (p *Poller) Latest() inav.RawIMU {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.imu
}
