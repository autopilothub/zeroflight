package inav_test

import (
	"context"
	"testing"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
)

func TestClientCloseIdempotent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := inav.DefaultConfig()
	cfg.Device = ""
	cfg.UDPAddress = "127.0.0.1:14550"

	client := inav.NewClient(cfg)
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}

	cancel()
	done := make(chan struct{})
	go func() {
		client.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("client close timed out")
	}

	// Second close must not panic.
	client.Close()
}
