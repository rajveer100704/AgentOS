# AgentOS Deployment & Production Resilience Guide

This document describes how to deploy the AgentOS platform in production-like environments using **Docker Compose** (for local observability) and **Kubernetes** (for highly resilient cloud-native setups).

---

## 1. Local Full-Stack Observability (Docker Compose)

AgentOS includes a complete observability suite using Docker Compose, integrating the Edge Proxy, Control Plane, Admin API, Prometheus, Grafana, and Jaeger.

### Prerequisites
*   Docker & Docker Compose installed.
*   The AgentOS binary built for Linux (if running within Docker containers) or run directly in hybrid-host mode.

### Compose Topology

```
                  [AI Client / User]
                          │
                          ▼
            [docker-compose: edge_proxy] (:8443)
                          │
             ┌────────────┴────────────┐
             ▼                         ▼
      [gateway] (:8080)         [admin_api] (:8081)
             │                         │
      (OTel Spans)               (OTel Spans)
             │                         │
             ▼                         ▼
       [jaeger] (:4317 OTLP) ◄── [prometheus] (scrapes :8081/metrics)
                                       │
                                       ▼
                               [grafana] (:3000)
```

### Deploying the Stack
1.  Configure the telemetry exporter to Jaeger in [configs/AgentOS.yaml](file:///c:/Users/BIT/AgentOS/AgentOS/configs/AgentOS.yaml):
    ```yaml
    telemetry:
      enabled: true
      exporter: "jaeger"
      otlp:
        endpoint: "jaeger:4317"
        insecure: true
    ```
2.  Start the compose stack:
    ```bash
    docker compose -f deployments/docker-compose.full.yaml up -d
    ```
3.  Access the interfaces:
    *   **Observatory Dashboard:** `http://localhost:8081` (directly served by Admin API)
    *   **Grafana Dashboards:** `http://localhost:3000` (default credentials: `admin/admin`)
    *   **Jaeger UI (Trace Explorer):** `http://localhost:16686`

---

## 2. Kubernetes Deployment (Production Resilience)

For production environments, AgentOS provides a set of declarative Kubernetes manifests that guarantee high availability, self-healing, and load-balancing.

Manifests are located in `deployments/k8s/`:
*   `deployment.yaml` — Deployment controller with resource constraints and health probes.
*   `service.yaml` — Service exposes Ports `:8080` (Gateway), `:8081` (Admin), `:8443` (Edge Proxy TLS), and `:8444` (Edge H2C).
*   `configmap.yaml` — Binds the `AgentOS.yaml` configurations into the containers.
*   `pdb.yaml` — Pod Disruption Budget ensuring HA during cluster upgrades.
*   `hpa.yaml` — Horizontal Pod Autoscaler scaling pods dynamically based on CPU utilization.

### Self-Healing & Health Probes
The deployment container defines `readinessProbe` and `livenessProbe` to ensure Kubernetes can automatically route traffic away from unhealthy pods and restart failed containers:

```yaml
# deployments/k8s/deployment.yaml
containers:
  - name: AgentOS
    image: AgentOS:latest
    readinessProbe:
      httpGet:
        path: /health
        port: 8081  # Admin API serves health metrics
      initialDelaySeconds: 5
      periodSeconds: 10
    livenessProbe:
      httpGet:
        path: /health
        port: 8081
      initialDelaySeconds: 15
      periodSeconds: 20
```

*   **Readiness Probe:** Runs after a 5-second initial delay. If the Admin API `/health` endpoint fails, the pod is removed from the Service endpoints, preventing clients from hitting a starting or deadlocked container.
*   **Liveness Probe:** Runs after a 15-second initial delay. If the pod fails this check, the kubelet kills the container and initiates a restart policy.

### Pod Disruption Budget (PDB)
To maintain structural resilience during node maintenance, drains, or upgrades, a PDB is defined to guarantee that at least 50% of the replicas remain active:

```yaml
# deployments/k8s/pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: AgentOS-pdb
  namespace: default
spec:
  minAvailable: 50%
  selector:
    matchLabels:
      app: AgentOS
```

If the deployment has 2 replicas (configured in `deployment.yaml`), this PDB ensures that Kubernetes will never evict more than 1 pod at a time during maintenance, preventing temporary service outages.

---

## 3. Prometheus & Grafana Configuration

The telemetry server exposes all metrics at `http://<admin-host>:8081/metrics`.

### Prometheus Scrape Job
Prometheus scrapes AgentOS metrics with the following configuration:

```yaml
# deployments/observability/prometheus.yml
scrape_configs:
  - job_name: 'AgentOS'
    static_configs:
      - targets: ['localhost:8081']
```

### Grafana Dashboard Panel Layout
The dashboard in `deployments/grafana/dashboards/AgentOS-edge.json` visualizes Edge and Control Plane metrics:
*   **Panel 1 (RPS):** Rates of requests forwarded to upstreams:
    `sum(rate(AgentOS_edge_requests_total[1m])) by (upstream)`
*   **Panel 2 (Connections):** Gauge of active edge connections:
    `AgentOS_edge_active_connections`
*   **Panel 3 (Load Shedding):** Counters of shed requests by reason:
    `sum(rate(AgentOS_loadshed_requests_total[1m])) by (reason)`
*   **Panel 4 (Circuit Breakers):** State timeline showing upstreams in Closed (0), Open (1), or Half-Open (2) states:
    `AgentOS_circuit_breaker_state`
*   **Panel 5 (LLM Latency):** Quantiles of LLM provider execution overhead:
    `histogram_quantile(0.95, sum(rate(AgentOS_provider_latency_seconds_bucket[5m])) by (le, provider))`
*   **Panel 6 (Traffic Control Overview):** Comparing **Served** vs **Rate-Limited** vs **Shed** request rates on a single timeline:
    *   *Query A (Served):* `sum(rate(AgentOS_edge_requests_total{status=~"2.."}[1m])) or sum(rate(AgentOS_telemetry_requests_total{status=~"2.."}[1m]))`
    *   *Query B (Rate-Limited):* `sum(rate(AgentOS_ratelimit_requests_total[1m]))`
    *   *Query C (Shed):* `sum(rate(AgentOS_loadshed_requests_total[1m]))`
