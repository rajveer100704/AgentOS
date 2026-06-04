# Observability

AgentOS exposes free local observability by default. No hosted monitoring service is required.

## Prometheus Metrics

Metrics are served from the admin port:

```bash
curl -fsS http://localhost:8081/metrics | grep '^AgentOS_'
```

Local Prometheus scrape config:

```bash
deployments/observability/prometheus.yml
```

The main gateway metrics are:

| Metric | Meaning |
|--------|---------|
| `AgentOS_requests_total` | Request count by tenant, method, path, and status |
| `AgentOS_request_duration_seconds` | Request latency histogram by tenant, method, and path |

## Grafana

Dashboard JSON:

```bash
deployments/grafana/AgentOS-dashboard.json
```

Import the dashboard into a local Grafana instance and point it at a local Prometheus datasource. The dashboard includes:

- gateway throughput
- p50 and p95 latency
- request count by status
- request count by tenant
- request count by path

## Local-Only Docker Example

Run AgentOS:

```bash
make demo-local
```

Run Prometheus locally with the included config:

```bash
docker run --rm -p 9090:9090 \
  -v "$PWD/deployments/observability/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  prom/prometheus:latest
```

Grafana can be run locally as well:

```bash
docker run --rm -p 3000:3000 grafana/grafana-oss:latest
```

These containers are optional. AgentOS itself exposes metrics without them.
