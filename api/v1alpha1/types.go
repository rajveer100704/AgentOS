package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AegisFlowGateway
type AegisFlowGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GatewaySpec   `json:"spec,omitempty"`
	Status            GatewayStatus `json:"status,omitempty"`
}

type GatewaySpec struct {
	Server    ServerSpec    `json:"server,omitempty"`
	Logging   LoggingSpec   `json:"logging,omitempty"`
	Telemetry TelemetrySpec `json:"telemetry,omitempty"`
	Cache     CacheSpec     `json:"cache,omitempty"`
	Database  DatabaseSpec  `json:"database,omitempty"`
	Admin     AdminSpec     `json:"admin,omitempty"`
}

type ServerSpec struct {
	Port      int `json:"port,omitempty"`
	AdminPort int `json:"adminPort,omitempty"`
}

type LoggingSpec struct {
	Level  string `json:"level,omitempty"`
	Format string `json:"format,omitempty"`
}

type TelemetrySpec struct {
	Enabled  bool   `json:"enabled,omitempty"`
	Exporter string `json:"exporter,omitempty"`
}

type CacheSpec struct {
	Enabled bool   `json:"enabled,omitempty"`
	Backend string `json:"backend,omitempty"`
	TTL     string `json:"ttl,omitempty"`
	MaxSize int    `json:"maxSize,omitempty"`
}

type DatabaseSpec struct {
	Enabled          bool      `json:"enabled,omitempty"`
	ConnStringSecret SecretRef `json:"connStringSecret,omitempty"`
}

type AdminSpec struct {
	TokenSecret SecretRef `json:"tokenSecret,omitempty"`
}

type SecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type GatewayStatus struct {
	Ready         bool   `json:"ready"`
	TotalRequests int64  `json:"totalRequests,omitempty"`
	ActiveTenants int    `json:"activeTenants,omitempty"`
	LastSyncedAt  string `json:"lastSyncedAt,omitempty"`
}

type AegisFlowGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisFlowGateway `json:"items"`
}

// AegisFlowProvider
type AegisFlowProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProviderSpec   `json:"spec,omitempty"`
	Status            ProviderStatus `json:"status,omitempty"`
}

type ProviderSpec struct {
	Type         string    `json:"type"`
	BaseURL      string    `json:"baseURL,omitempty"`
	APIKeySecret SecretRef `json:"apiKeySecret,omitempty"`
	Models       []string  `json:"models,omitempty"`
	Timeout      string    `json:"timeout,omitempty"`
	MaxRetries   int       `json:"maxRetries,omitempty"`
	Region       string    `json:"region,omitempty"`
}

type ProviderStatus struct {
	Healthy          bool    `json:"healthy"`
	AverageLatencyMs int64   `json:"averageLatencyMs,omitempty"`
	ErrorRate        float64 `json:"errorRate,omitempty"`
	LastCheckedAt    string  `json:"lastCheckedAt,omitempty"`
}

type AegisFlowProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisFlowProvider `json:"items"`
}

// AegisFlowRoute
type AegisFlowRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RouteSpec   `json:"spec,omitempty"`
	Status            RouteStatus `json:"status,omitempty"`
}

type RouteSpec struct {
	Match   RouteMatchSpec `json:"match"`
	Regions []RouteRegion  `json:"regions,omitempty"`
	Canary  *CanarySpec    `json:"canary,omitempty"`
}

type RouteMatchSpec struct {
	Model string `json:"model"`
}

type RouteRegion struct {
	Name      string   `json:"name"`
	Providers []string `json:"providers"`
	Strategy  string   `json:"strategy,omitempty"`
}

type CanarySpec struct {
	TargetProvider      string  `json:"targetProvider"`
	Stages              []int   `json:"stages"`
	ObservationWindow   string  `json:"observationWindow"`
	ErrorThreshold      float64 `json:"errorThreshold"`
	LatencyP95Threshold int64   `json:"latencyP95Threshold"`
}

type RouteStatus struct {
	ActiveCanary     bool   `json:"activeCanary,omitempty"`
	CanaryPercentage int    `json:"canaryPercentage,omitempty"`
	CanaryHealth     string `json:"canaryHealth,omitempty"`
}

type AegisFlowRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisFlowRoute `json:"items"`
}

// AegisFlowTenant
type AegisFlowTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TenantSpec   `json:"spec,omitempty"`
	Status            TenantStatus `json:"status,omitempty"`
}

type TenantSpec struct {
	DisplayName   string            `json:"displayName,omitempty"`
	APIKeySecrets []SecretRef       `json:"apiKeySecrets,omitempty"`
	RateLimit     RateLimitSpec     `json:"rateLimit,omitempty"`
	AllowedModels []string          `json:"allowedModels,omitempty"`
	Budget        *TenantBudgetSpec `json:"budget,omitempty"`
}

type RateLimitSpec struct {
	RequestsPerMinute int `json:"requestsPerMinute,omitempty"`
	TokensPerMinute   int `json:"tokensPerMinute,omitempty"`
}

type TenantBudgetSpec struct {
	Monthly float64                    `json:"monthly,omitempty"`
	AlertAt int                        `json:"alertAt,omitempty"`
	WarnAt  int                        `json:"warnAt,omitempty"`
	Models  map[string]ModelBudgetSpec `json:"models,omitempty"`
}

type ModelBudgetSpec struct {
	Monthly float64 `json:"monthly,omitempty"`
	AlertAt int     `json:"alertAt,omitempty"`
	WarnAt  int     `json:"warnAt,omitempty"`
}

type TenantStatus struct {
	TotalRequests        int64   `json:"totalRequests,omitempty"`
	TotalTokens          int64   `json:"totalTokens,omitempty"`
	EstimatedCost        float64 `json:"estimatedCost,omitempty"`
	BudgetPercentage     float64 `json:"budgetPercentage,omitempty"`
	RateLimitUtilization int     `json:"rateLimitUtilization,omitempty"`
}

type AegisFlowTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisFlowTenant `json:"items"`
}

// AegisFlowPolicy
type AegisFlowPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PolicySpec `json:"spec,omitempty"`
}

type PolicySpec struct {
	Phase    string   `json:"phase"`
	Type     string   `json:"type"`
	Action   string   `json:"action"`
	Keywords []string `json:"keywords,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
	WasmPath string   `json:"wasmPath,omitempty"`
	Timeout  string   `json:"timeout,omitempty"`
	OnError  string   `json:"onError,omitempty"`
}

type AegisFlowPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisFlowPolicy `json:"items"`
}
