package inav

import (
	"context"
	"fmt"
	"time"

	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

type missionTransfer struct {
	waypoints []mission.Waypoint
	done      chan error
}

// ClearMission removes all stored mission items on the FC.
func (c *Client) ClearMission(ctx context.Context) error {
	if err := c.sendMissionClearAll(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(300 * time.Millisecond):
		return nil
	}
}

// UploadMission uploads waypoints to INAV using the MAVLink mission protocol.
// The vehicle must be disarmed.
func (c *Client) UploadMission(ctx context.Context, waypoints []mission.Waypoint) error {
	if c.State().Armed {
		return fmt.Errorf("cannot upload mission while armed")
	}
	if len(waypoints) == 0 {
		return fmt.Errorf("mission has no waypoints")
	}

	c.mu.Lock()
	if c.missionXfer != nil {
		c.mu.Unlock()
		return fmt.Errorf("mission transfer already in progress")
	}
	xfer := &missionTransfer{
		waypoints: waypoints,
		done:      make(chan error, 1),
	}
	c.missionXfer = xfer
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.missionXfer = nil
		c.mu.Unlock()
	}()

	if err := c.sendMissionClearAll(); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	if err := c.sendMissionCount(uint16(len(waypoints))); err != nil {
		return err
	}

	select {
	case err := <-xfer.done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timed out waiting for mission upload acknowledgement")
	}
}

func (c *Client) sendMissionClearAll() error {
	c.mu.RLock()
	node := c.node
	channel := c.channel
	c.mu.RUnlock()
	if node == nil || channel == nil {
		return fmt.Errorf("mavlink not connected")
	}
	return node.WriteMessageTo(channel, &ardupilotmega.MessageMissionClearAll{
		TargetSystem:    c.cfg.TargetSystemID,
		TargetComponent: c.cfg.TargetComponentID,
		MissionType:     ardupilotmega.MAV_MISSION_TYPE_MISSION,
	})
}

func (c *Client) sendMissionCount(count uint16) error {
	c.mu.RLock()
	node := c.node
	channel := c.channel
	c.mu.RUnlock()
	if node == nil || channel == nil {
		return fmt.Errorf("mavlink not connected")
	}
	return node.WriteMessageTo(channel, &ardupilotmega.MessageMissionCount{
		TargetSystem:    c.cfg.TargetSystemID,
		TargetComponent: c.cfg.TargetComponentID,
		Count:           count,
		MissionType:     ardupilotmega.MAV_MISSION_TYPE_MISSION,
	})
}

func (c *Client) handleMissionRequest(channel *gomavlib.Channel, seq uint16) error {
	c.mu.Lock()
	xfer := c.missionXfer
	c.mu.Unlock()
	if xfer == nil {
		return nil
	}
	if int(seq) >= len(xfer.waypoints) {
		select {
		case xfer.done <- fmt.Errorf("requested mission seq %d out of range", seq):
		default:
		}
		return fmt.Errorf("mission seq %d out of range", seq)
	}

	item := toMissionItem(seq, xfer.waypoints[seq], c.cfg)
	c.mu.RLock()
	node := c.node
	c.mu.RUnlock()
	if node == nil {
		return fmt.Errorf("mavlink node is not initialized")
	}
	return node.WriteMessageTo(channel, item)
}

func (c *Client) handleMissionAck(msg *ardupilotmega.MessageMissionAck) {
	c.mu.Lock()
	xfer := c.missionXfer
	c.mu.Unlock()
	if xfer == nil {
		return
	}

	var err error
	if msg.Type != ardupilotmega.MAV_MISSION_ACCEPTED {
		err = fmt.Errorf("mission upload rejected: %s", msg.Type)
	}
	select {
	case xfer.done <- err:
	default:
	}
}

func toMissionItem(seq uint16, wp mission.Waypoint, cfg Config) *ardupilotmega.MessageMissionItem {
	return &ardupilotmega.MessageMissionItem{
		TargetSystem:    cfg.TargetSystemID,
		TargetComponent: cfg.TargetComponentID,
		Seq:             seq,
		Frame:           ardupilotmega.MAV_FRAME_GLOBAL,
		Command:         common.MAV_CMD(ardupilotmega.MAV_CMD_NAV_WAYPOINT),
		Autocontinue:    1,
		X:               float32(wp.Lat),
		Y:               float32(wp.Lon),
		Z:               wp.Alt,
		MissionType:     ardupilotmega.MAV_MISSION_TYPE_MISSION,
	}
}

func (c *Client) handleMissionFrame(channel *gomavlib.Channel, msg any) {
	switch m := msg.(type) {
	case *ardupilotmega.MessageMissionRequest:
		if err := c.handleMissionRequest(channel, m.Seq); err != nil {
			c.failMissionTransfer(err)
		}
	case *ardupilotmega.MessageMissionRequestInt:
		if err := c.handleMissionRequest(channel, m.Seq); err != nil {
			c.failMissionTransfer(err)
		}
	case *ardupilotmega.MessageMissionAck:
		c.handleMissionAck(m)
	}
}

func (c *Client) failMissionTransfer(err error) {
	c.mu.Lock()
	xfer := c.missionXfer
	c.mu.Unlock()
	if xfer == nil {
		return
	}
	select {
	case xfer.done <- err:
	default:
	}
}
