package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan is the single entry point for all span creation in AegisFlow.
// It wraps otel.Tracer().Start() with the AegisFlow tracer and returns the
// enriched context and span.
//
// Usage in gateway/handler.go:
//
//	ctx, span := telemetry.StartSpan(ctx, "policy.check_input",
//	    attribute.String(telemetry.AttrTenantID, tenantID),
//	)
//	defer span.End()
//	violation, err := h.policy.CheckInput(content)
//	if err != nil {
//	    telemetry.RecordError(span, err)
//	}
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
}

// RecordError marks the span as errored and records the error event.
// No-op if span is nil or not recording.
func RecordError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// SetDecision sets the policy decision attribute on a span.
// decision should be one of "allow", "review", "block", "warn".
func SetDecision(span trace.Span, decision string) {
	if span == nil {
		return
	}
	span.SetAttributes(attribute.String(AttrDecision, decision))
}

// SetRiskScore sets the behavioral risk score attribute on a span.
func SetRiskScore(span trace.Span, score int) {
	if span == nil {
		return
	}
	span.SetAttributes(attribute.Int(AttrRiskScore, score))
}

// SetProvider sets provider-related attributes on a span.
func SetProvider(span trace.Span, provider, model string) {
	if span == nil {
		return
	}
	span.SetAttributes(
		attribute.String(AttrProvider, provider),
		attribute.String(AttrModel, model),
	)
}

// SetTokens records token usage attributes on a provider span.
func SetTokens(span trace.Span, prompt, completion, total int) {
	if span == nil {
		return
	}
	span.SetAttributes(
		attribute.Int(AttrTokensPrompt, prompt),
		attribute.Int(AttrTokensCompletion, completion),
		attribute.Int(AttrTokensTotal, total),
	)
}

// SpanFromContext extracts the current span from a context.
// Returns trace.SpanFromContext — never nil (may be no-op).
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}
