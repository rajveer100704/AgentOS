package rollout

import "time"

const (
	StatePending    = "pending"
	StateRunning    = "running"
	StatePaused     = "paused"
	StateRolledBack = "rolled_back"
	StateCompleted  = "completed"
)

type Rollout struct {
	ID                  string        `json:"id"`
	RouteModel          string        `json:"route_model"`
	BaselineProviders   []string      `json:"baseline_providers"`
	CanaryProvider      string        `json:"canary_provider"`
	Stages              []int         `json:"stages"`
	CurrentStage        int           `json:"current_stage"`
	CurrentPercentage   int           `json:"current_percentage"`
	State               string        `json:"state"`
	ObservationWindow   time.Duration `json:"observation_window"`
	ErrorThreshold      float64       `json:"error_threshold"`
	LatencyP95Threshold int64         `json:"latency_p95_threshold"`
	StageStartedAt      time.Time     `json:"stage_started_at"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
	CompletedAt         *time.Time    `json:"completed_at,omitempty"`
	RollbackReason      string        `json:"rollback_reason,omitempty"`
}

type HealthMetrics struct {
	ErrorRate    float64 `json:"error_rate"`
	P95LatencyMs int64   `json:"p95_latency_ms"`
	Requests     int64   `json:"requests"`
}

type RolloutMetrics struct {
	Baseline HealthMetrics `json:"baseline"`
	Canary   HealthMetrics `json:"canary"`
}

type Store interface {
	Create(r *Rollout) error
	Get(id string) (*Rollout, error)
	GetByModel(model string) (*Rollout, error)
	Update(r *Rollout) error
	List() ([]*Rollout, error)
	Migrate() error
}
