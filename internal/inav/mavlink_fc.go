package inav

import (
	"context"
	"time"
)

// MAVLinkAdapter wraps Client for the link.FC interface.
type MAVLinkAdapter struct {
	*Client
}

// Close shuts down the MAVLink client.
func (a *MAVLinkAdapter) Close() error {
	a.Client.Close()
	return nil
}

// WaitReady blocks until MAVLink telemetry is available.
func (a *MAVLinkAdapter) WaitReady(ctx context.Context, timeout time.Duration) error {
	return a.Client.WaitForConnection(ctx, timeout)
}
