package inav

import (
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

func (c *Client) requestTelemetryStreams(channel *gomavlib.Channel) {
	if c.node == nil || channel == nil {
		return
	}

	streams := []struct {
		id   ardupilotmega.MAV_DATA_STREAM
		rate uint16
	}{
		{ardupilotmega.MAV_DATA_STREAM_EXTRA1, 10},           // ATTITUDE
		{ardupilotmega.MAV_DATA_STREAM_POSITION, 10},       // GPS
		{ardupilotmega.MAV_DATA_STREAM_EXTENDED_STATUS, 2}, // SYS_STATUS
		{ardupilotmega.MAV_DATA_STREAM_RC_CHANNELS, 2},
	}

	for _, stream := range streams {
		msg := &ardupilotmega.MessageRequestDataStream{
			TargetSystem:    c.cfg.TargetSystemID,
			TargetComponent: c.cfg.TargetComponentID,
			ReqStreamId:     stream.id,
			ReqMessageRate:  stream.rate,
			StartStop:       1,
		}
		_ = c.node.WriteMessageTo(channel, msg)
	}
}
