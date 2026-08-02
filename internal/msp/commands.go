package msp

// MSP v1 command IDs used with INAV.
const (
	CmdSTATUS       = 101
	CmdRAWIMU       = 102
	CmdRAWGPS       = 106
	CmdATTITUDE     = 108
	CmdALTITUDE     = 109
	CmdANALOG       = 110
	CmdACTIVEBOXES  = 113
	CmdWPCOUNT      = 44
	CmdWP           = 118
	CmdBOXIDS       = 119
	CmdNAVSTATUS    = 121
	CmdSETWP        = 209
)

// INAV permanent box IDs (subset).
const (
	BoxArm    = 0
	BoxGCSNav = 31
)
