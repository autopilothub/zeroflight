package link

import (
	"context"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/internal/mission"
)

// FC is the flight-controller link used by CLI and API.
type FC interface {
	State() inav.VehicleState
	Close() error
	WaitReady(ctx context.Context, timeout time.Duration) error
	SendGoto(req inav.GotoRequest) error
	UploadMission(ctx context.Context, waypoints []mission.Waypoint) error
	ClearMission(ctx context.Context) error
}
