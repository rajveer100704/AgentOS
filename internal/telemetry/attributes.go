package telemetry

// AegisFlow semantic attribute keys for OpenTelemetry spans.
// Attribute values are set via telemetry.StartSpan() or span.SetAttributes().

const (
	// Request attributes
	AttrTenantID = "aegisflow.tenant.id"
	AttrModel    = "aegisflow.model"
	AttrProvider = "aegisflow.provider"
	AttrStream   = "aegisflow.stream"

	// Token usage attributes (on provider spans)
	AttrTokensPrompt     = "aegisflow.tokens.prompt"
	AttrTokensCompletion = "aegisflow.tokens.completion"
	AttrTokensTotal      = "aegisflow.tokens.total"

	// Cost
	AttrCostUSD = "aegisflow.cost.usd"

	// Policy attributes
	AttrPolicyViolated = "aegisflow.policy.violated"
	AttrPolicyName     = "aegisflow.policy.name"
	AttrDecision       = "aegisflow.decision"

	// Risk / behavioral
	AttrRiskScore = "aegisflow.risk.score"
	AttrSessionID = "aegisflow.session.id"

	// Evidence
	AttrEvidenceHash = "aegisflow.evidence.hash"

	// Credential
	AttrCredentialName = "aegisflow.credential.name"
	AttrCredentialTTL  = "aegisflow.credential.ttl_seconds"

	// Edge proxy
	AttrEdgeUpstream   = "aegisflow.edge.upstream"
	AttrEdgeCPUPercent = "aegisflow.edge.cpu_pct"

	// Approval
	AttrApprovalID = "aegisflow.approval.id"
)
