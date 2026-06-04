package analytics

import (
	"fmt"
	"math"
	"time"
)

type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityWarning  AlertSeverity = "warning"
	SeverityInfo     AlertSeverity = "info"
)

type Alert struct {
	ID         string        `json:"id"`
	Severity   AlertSeverity `json:"severity"`
	Type       string        `json:"type"`
	Dimension  string        `json:"dimension"`
	Metric     string        `json:"metric"`
	Value      float64       `json:"value"`
	Threshold  float64       `json:"threshold"`
	Message    string        `json:"message"`
	State      string        `json:"state"`
	CreatedAt  time.Time     `json:"created_at"`
	ResolvedAt *time.Time    `json:"resolved_at,omitempty"`
}

type DetectionResult struct {
	Alerts []Alert
}

type Detector struct {
	collector      *Collector
	staticConfig   StaticThresholds
	baselineConfig BaselineConfig
}

// StaticThresholds and BaselineConfig are from config package, but to avoid import
// we duplicate the needed fields here. The caller maps from config types.
type StaticThresholds struct {
	ErrorRateMax         float64
	P95LatencyMax        int64
	RequestsPerMinuteMax int64
	CostPerMinuteMax     float64
}

type BaselineConfig struct {
	Window          time.Duration
	StddevThreshold float64
}

func NewDetector(collector *Collector, static StaticThresholds, baseline BaselineConfig) *Detector {
	return &Detector{collector: collector, staticConfig: static, baselineConfig: baseline}
}

func (d *Detector) Evaluate() DetectionResult {
	var alerts []Alert
	for _, dim := range d.collector.Dimensions() {
		ts := d.collector.GetSeries(dim)
		if ts == nil {
			continue
		}
		alerts = append(alerts, d.checkStatic(dim, ts)...)
		alerts = append(alerts, d.checkBaseline(dim, ts)...)
	}
	return DetectionResult{Alerts: alerts}
}

func (d *Detector) checkStatic(dim string, ts *TimeSeries) []Alert {
	// Check last 1-minute bucket
	buckets := ts.RecentBuckets(1)
	if len(buckets) == 0 {
		return nil
	}
	b := buckets[0]
	var alerts []Alert

	if b.ErrorRate > d.staticConfig.ErrorRateMax && d.staticConfig.ErrorRateMax > 0 {
		alerts = append(alerts, Alert{
			ID:        fmt.Sprintf("static-%s-error_rate-%d", dim, time.Now().UnixNano()),
			Severity:  SeverityCritical,
			Type:      "static_threshold",
			Dimension: dim,
			Metric:    "error_rate",
			Value:     b.ErrorRate,
			Threshold: d.staticConfig.ErrorRateMax,
			Message:   fmt.Sprintf("%s error rate %.1f%% exceeds threshold %.1f%%", dim, b.ErrorRate, d.staticConfig.ErrorRateMax),
			State:     "active",
			CreatedAt: time.Now(),
		})
	}

	if b.P95Latency > d.staticConfig.P95LatencyMax && d.staticConfig.P95LatencyMax > 0 {
		alerts = append(alerts, Alert{
			ID:        fmt.Sprintf("static-%s-p95_latency-%d", dim, time.Now().UnixNano()),
			Severity:  SeverityCritical,
			Type:      "static_threshold",
			Dimension: dim,
			Metric:    "p95_latency",
			Value:     float64(b.P95Latency),
			Threshold: float64(d.staticConfig.P95LatencyMax),
			Message:   fmt.Sprintf("%s p95 latency %dms exceeds threshold %dms", dim, b.P95Latency, d.staticConfig.P95LatencyMax),
			State:     "active",
			CreatedAt: time.Now(),
		})
	}

	if b.Requests > d.staticConfig.RequestsPerMinuteMax && d.staticConfig.RequestsPerMinuteMax > 0 {
		alerts = append(alerts, Alert{
			ID:        fmt.Sprintf("static-%s-request_rate-%d", dim, time.Now().UnixNano()),
			Severity:  SeverityCritical,
			Type:      "static_threshold",
			Dimension: dim,
			Metric:    "request_rate",
			Value:     float64(b.Requests),
			Threshold: float64(d.staticConfig.RequestsPerMinuteMax),
			Message:   fmt.Sprintf("%s request rate %d/min exceeds threshold %d/min", dim, b.Requests, d.staticConfig.RequestsPerMinuteMax),
			State:     "active",
			CreatedAt: time.Now(),
		})
	}

	if b.EstimatedCost > d.staticConfig.CostPerMinuteMax && d.staticConfig.CostPerMinuteMax > 0 {
		alerts = append(alerts, Alert{
			ID:        fmt.Sprintf("static-%s-cost_rate-%d", dim, time.Now().UnixNano()),
			Severity:  SeverityCritical,
			Type:      "static_threshold",
			Dimension: dim,
			Metric:    "cost_rate",
			Value:     b.EstimatedCost,
			Threshold: d.staticConfig.CostPerMinuteMax,
			Message:   fmt.Sprintf("%s cost $%.2f/min exceeds threshold $%.2f/min", dim, b.EstimatedCost, d.staticConfig.CostPerMinuteMax),
			State:     "active",
			CreatedAt: time.Now(),
		})
	}

	return alerts
}

func (d *Detector) checkBaseline(dim string, ts *TimeSeries) []Alert {
	baselineMinutes := int(d.baselineConfig.Window.Minutes())
	allBuckets := ts.RecentBuckets(baselineMinutes)
	if len(allBuckets) < 60 { // need at least 1h of data
		return nil
	}

	// Baseline = all except last 5 minutes
	if len(allBuckets) <= 5 {
		return nil
	}
	baseline := allBuckets[:len(allBuckets)-5]
	recent := allBuckets[len(allBuckets)-5:]

	var alerts []Alert

	// Check request rate anomaly
	if alert := d.checkMetricAnomaly(dim, "request_rate", baseline, recent,
		func(b BucketSummary) float64 { return float64(b.Requests) }); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Check error rate anomaly
	if alert := d.checkMetricAnomaly(dim, "error_rate", baseline, recent,
		func(b BucketSummary) float64 { return b.ErrorRate }); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Check p95 latency anomaly
	if alert := d.checkMetricAnomaly(dim, "p95_latency", baseline, recent,
		func(b BucketSummary) float64 { return float64(b.P95Latency) }); alert != nil {
		alerts = append(alerts, *alert)
	}

	return alerts
}

func (d *Detector) checkMetricAnomaly(dim, metric string, baseline, recent []BucketSummary, extract func(BucketSummary) float64) *Alert {
	// Calculate baseline mean and stddev
	var sum, sumSq float64
	n := float64(len(baseline))
	for _, b := range baseline {
		v := extract(b)
		sum += v
		sumSq += v * v
	}
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	stddev := math.Sqrt(variance)

	// Calculate recent average
	var recentSum float64
	for _, b := range recent {
		recentSum += extract(b)
	}
	recentAvg := recentSum / float64(len(recent))

	// Check deviation
	if stddev > 0 && math.Abs(recentAvg-mean) > stddev*d.baselineConfig.StddevThreshold {
		return &Alert{
			ID:        fmt.Sprintf("baseline-%s-%s-%d", dim, metric, time.Now().UnixNano()),
			Severity:  SeverityWarning,
			Type:      "statistical_baseline",
			Dimension: dim,
			Metric:    metric,
			Value:     recentAvg,
			Threshold: mean + stddev*d.baselineConfig.StddevThreshold,
			Message:   fmt.Sprintf("%s %s %.1f deviates from baseline %.1f (±%.1f, threshold: %.1f stddev)", dim, metric, recentAvg, mean, stddev, d.baselineConfig.StddevThreshold),
			State:     "active",
			CreatedAt: time.Now(),
		}
	}
	return nil
}
