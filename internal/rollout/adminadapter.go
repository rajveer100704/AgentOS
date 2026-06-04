package rollout

import "time"

// AdminAdapter wraps a *Manager and satisfies the admin.RolloutManager interface
// without creating an import cycle. All concrete rollout types are returned as
// any so the admin package can JSON-encode them without importing this package.
type AdminAdapter struct {
	mgr *Manager
}

// NewAdminAdapter creates an AdminAdapter for the given Manager.
func NewAdminAdapter(m *Manager) *AdminAdapter {
	return &AdminAdapter{mgr: m}
}

func (a *AdminAdapter) ListRollouts() (any, error) {
	return a.mgr.ListRollouts()
}

func (a *AdminAdapter) CreateRollout(
	routeModel string,
	baselineProviders []string,
	canaryProvider string,
	stages []int,
	observationWindow time.Duration,
	errorThreshold float64,
	latencyP95Threshold int64,
) (any, error) {
	return a.mgr.CreateRollout(routeModel, baselineProviders, canaryProvider, stages, observationWindow, errorThreshold, latencyP95Threshold)
}

func (a *AdminAdapter) GetRolloutWithMetrics(id string) (any, error) {
	r, err := a.mgr.GetRollout(id)
	if err != nil {
		return nil, err
	}
	metrics := a.mgr.GetMetrics(r)
	return map[string]any{
		"rollout": r,
		"metrics": metrics,
	}, nil
}

func (a *AdminAdapter) PauseRollout(id string) error {
	return a.mgr.PauseRollout(id)
}

func (a *AdminAdapter) ResumeRollout(id string) error {
	return a.mgr.ResumeRollout(id)
}

func (a *AdminAdapter) RollbackRollout(id string) error {
	return a.mgr.RollbackRollout(id)
}
