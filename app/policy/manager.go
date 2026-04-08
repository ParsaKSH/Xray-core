package policy

import (
	"context"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/features/policy"
)

// Instance is an instance of Policy manager.
type Instance struct {
	levels map[uint32]*Policy
	system *SystemPolicy
}

// New creates new Policy manager instance.
func New(ctx context.Context, config *Config) (*Instance, error) {
	m := &Instance{
		levels: make(map[uint32]*Policy),
		system: config.System,
	}
	if len(config.Level) > 0 {
		for lv, p := range config.Level {
			pp := defaultPolicy()
			pp.overrideWith(p)
			// Read speed limit from the package-level registry
			// (set by infra/conf during JSON parsing, bypasses protobuf serialization)
			if sl := GetLevelSpeedLimit(lv); sl > 0 {
				pp.SpeedLimit = sl
			}
			errors.LogInfo(ctx, "[policy] level ", lv, " speedLimit=", pp.SpeedLimit, " bytes/sec")
			m.levels[lv] = pp
		}
	}

	return m, nil
}

// Type implements common.HasType.
func (*Instance) Type() interface{} {
	return policy.ManagerType()
}

// ForLevel implements policy.Manager.
func (m *Instance) ForLevel(level uint32) policy.Session {
	if p, ok := m.levels[level]; ok {
		s := p.ToCorePolicy()
		if s.SpeedLimit > 0 {
			errors.LogInfo(context.Background(), "[policy] ForLevel(", level, ") returning SpeedLimit=", s.SpeedLimit)
		}
		return s
	}
	return policy.SessionDefault()
}

// ForSystem implements policy.Manager.
func (m *Instance) ForSystem() policy.System {
	if m.system == nil {
		return policy.System{}
	}
	return m.system.ToCorePolicy()
}

// Start implements common.Runnable.Start().
func (m *Instance) Start() error {
	return nil
}

// Close implements common.Closable.Close().
func (m *Instance) Close() error {
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
