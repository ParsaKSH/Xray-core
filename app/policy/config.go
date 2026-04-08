package policy

import (
	"sync"
	"time"

	"github.com/xtls/xray-core/features/policy"
)

// LevelSpeedLimits stores speed limits (bytes/sec) per policy level.
// Set by infra/conf during config parsing, read by the policy manager.
var (
	LevelSpeedLimits   = make(map[uint32]uint64)
	LevelSpeedLimitsMu sync.Mutex
)

// SetLevelSpeedLimit sets the speed limit for a given policy level.
func SetLevelSpeedLimit(level uint32, bytesPerSec uint64) {
	LevelSpeedLimitsMu.Lock()
	defer LevelSpeedLimitsMu.Unlock()
	LevelSpeedLimits[level] = bytesPerSec
}

// GetLevelSpeedLimit returns the speed limit for a given policy level.
func GetLevelSpeedLimit(level uint32) uint64 {
	LevelSpeedLimitsMu.Lock()
	defer LevelSpeedLimitsMu.Unlock()
	return LevelSpeedLimits[level]
}

// Duration converts Second to time.Duration.
func (s *Second) Duration() time.Duration {
	if s == nil {
		return 0
	}
	return time.Second * time.Duration(s.Value)
}

func defaultPolicy() *Policy {
	p := policy.SessionDefault()

	return &Policy{
		Timeout: &Policy_Timeout{
			Handshake:      &Second{Value: uint32(p.Timeouts.Handshake / time.Second)},
			ConnectionIdle: &Second{Value: uint32(p.Timeouts.ConnectionIdle / time.Second)},
			UplinkOnly:     &Second{Value: uint32(p.Timeouts.UplinkOnly / time.Second)},
			DownlinkOnly:   &Second{Value: uint32(p.Timeouts.DownlinkOnly / time.Second)},
		},
		Buffer: &Policy_Buffer{
			Connection: p.Buffer.PerConnection,
		},
	}
}

func (p *Policy_Timeout) overrideWith(another *Policy_Timeout) {
	if another.Handshake != nil {
		p.Handshake = &Second{Value: another.Handshake.Value}
	}
	if another.ConnectionIdle != nil {
		p.ConnectionIdle = &Second{Value: another.ConnectionIdle.Value}
	}
	if another.UplinkOnly != nil {
		p.UplinkOnly = &Second{Value: another.UplinkOnly.Value}
	}
	if another.DownlinkOnly != nil {
		p.DownlinkOnly = &Second{Value: another.DownlinkOnly.Value}
	}
}

func (p *Policy) overrideWith(another *Policy) {
	if another.Timeout != nil {
		p.Timeout.overrideWith(another.Timeout)
	}
	if another.Stats != nil && p.Stats == nil {
		p.Stats = &Policy_Stats{}
		p.Stats = another.Stats
	}
	if another.Buffer != nil {
		p.Buffer = &Policy_Buffer{
			Connection: another.Buffer.Connection,
		}
	}
	if another.SpeedLimit > 0 {
		p.SpeedLimit = another.SpeedLimit
	}
}

// ToCorePolicy converts this Policy to policy.Session.
func (p *Policy) ToCorePolicy() policy.Session {
	cp := policy.SessionDefault()

	if p.Timeout != nil {
		cp.Timeouts.ConnectionIdle = p.Timeout.ConnectionIdle.Duration()
		cp.Timeouts.Handshake = p.Timeout.Handshake.Duration()
		cp.Timeouts.DownlinkOnly = p.Timeout.DownlinkOnly.Duration()
		cp.Timeouts.UplinkOnly = p.Timeout.UplinkOnly.Duration()
	}
	if p.Stats != nil {
		cp.Stats.UserUplink = p.Stats.UserUplink
		cp.Stats.UserDownlink = p.Stats.UserDownlink
		cp.Stats.UserOnline = p.Stats.UserOnline
	}
	if p.Buffer != nil {
		cp.Buffer.PerConnection = p.Buffer.Connection
	}
	cp.SpeedLimit = p.SpeedLimit
	return cp
}

// ToCorePolicy converts this SystemPolicy to policy.System.
func (p *SystemPolicy) ToCorePolicy() policy.System {
	return policy.System{
		Stats: policy.SystemStats{
			InboundUplink:    p.Stats.InboundUplink,
			InboundDownlink:  p.Stats.InboundDownlink,
			OutboundUplink:   p.Stats.OutboundUplink,
			OutboundDownlink: p.Stats.OutboundDownlink,
		},
	}
}
