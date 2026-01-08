# PodWatch

**Kubernetes Runtime Threat Detection and Response System**

PodWatch monitors live container behavior, detects malicious runtime activity in under 5 seconds, creates incidents with timelines, and executes automated response playbooks with guardrails.

## Features

- **Real-time Detection**: Alerts within 5 seconds of malicious activity
- **Falco Integration**: Uses Falco as a DaemonSet for runtime telemetry
- **CEL-based Rules**: Flexible rule engine using Common Expression Language
- **Automated Response**: Configurable playbooks with safety guardrails
- **Incident Management**: Full timeline tracking with evidence bundles
- **Modern UI**: React-based dashboard for alerts, incidents, and rules

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐   │
│  │  Falco   │───▶│  Ingest  │───▶│  Enrich  │───▶│  Detect  │   │
│  │ DaemonSet│    │   API    │    │ Service  │    │  Engine  │   │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘   │
│       │               │               │               │          │
│       │               ▼               ▼               ▼          │
│       │          ┌─────────┐    ┌─────────┐    ┌──────────┐     │
│       │          │  MinIO  │    │  NATS   │    │  Redis   │     │
│       │          │   S3    │    │JetStream│    │ Correlate│     │
│       │          └─────────┘    └─────────┘    └──────────┘     │
│       │                                              │           │
│       │                                              ▼           │
│       │          ┌──────────┐    ┌──────────┐    ┌──────────┐   │
│       │          │ Incident │◀───│  Alert   │◀───│ Response │   │
│       │          │ Service  │    │  Store   │    │Orchestr. │   │
│       │          └──────────┘    └──────────┘    └──────────┘   │
│       │               │                               │          │
│       │               ▼                               ▼          │
│       │          ┌─────────┐                   ┌──────────┐     │
│       │          │Postgres │                   │   K8s    │     │
│       │          │   DB    │                   │   API    │     │
│       │          └─────────┘                   └──────────┘     │
│       │                                                          │
│       │          ┌──────────────────────────────────────┐       │
│       └─────────▶│              Web UI                   │       │
│                  │   Alerts | Incidents | Rules | Replay │       │
│                  └──────────────────────────────────────┘       │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Sensor | Falco DaemonSet |
| Backend | Go |
| Event Pipeline | NATS JetStream |
| Storage | PostgreSQL, OpenSearch, MinIO |
| State | Redis |
| UI | React + TypeScript |
| Deployment | Helm |

## Repository Structure

```
podwatch/
├── sensor/          # Falco configuration
├── ingest/          # Event ingestion API
├── enrich/          # K8s metadata enrichment
├── detect/          # Detection engine
│   ├── matcher/     # CEL rule matcher
│   ├── correlator/  # Redis-based correlation
│   └── rules/       # Default rule definitions
├── incident/        # Incident management service
├── respond/         # Response orchestrator
├── ui/              # React web interface
├── deploy/          # Deployment manifests
│   └── helm/        # Helm charts
├── test/            # Tests and attack scripts
│   ├── attacks/     # Attack simulations
│   └── fixtures/    # Test event fixtures
└── docs/            # Documentation
```

## Quick Start

### Prerequisites

- [kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/)
- [Go 1.21+](https://golang.org/)
- [Node.js 18+](https://nodejs.org/)

### 1. Create Kind Cluster

```bash
kind create cluster --name kubeguard
```

### 2. Deploy PodWatch

```bash
cd deploy/helm
helm install kubeguard ./kubeguard -n security-system --create-namespace
```

### 3. Wait for Components

```bash
kubectl get pods -n security-system -w
```

### 4. Access UI

```bash
kubectl port-forward svc/kubeguard-ui 3000:80 -n security-system
```

Open http://localhost:3000

### 5. Run Attack Simulations

```bash
cd test/attacks
chmod +x *.sh
./run_all.sh
```

## Detection Rules

PodWatch uses CEL (Common Expression Language) for rule definitions:

```yaml
rules:
  - id: "rule-shell-spawn"
    name: "Shell Spawn in Production"
    severity: "high"
    condition: |
      event.process.exe in ['/bin/bash', '/bin/sh'] && 
      event.container.namespace == 'prod'
    response: "kill_pod"
    enabled: true
```

### Built-in Rules

| Rule | Severity | Response |
|------|----------|----------|
| Shell Spawn in Prod | High | Kill Pod |
| Service Account Token Read | High | Quarantine Namespace |
| Reverse Shell Indicators | Critical | Kill Pod + Ticket |
| Privilege Escalation | Critical | Isolate Node |
| Package Manager in Prod | Medium | Alert Only |

## Response Actions

### Available Playbooks

1. **Kill Pod** - Immediately terminate the malicious pod
2. **Quarantine Namespace** - Apply NetworkPolicy to isolate namespace
3. **Isolate Node** - Cordon the node to prevent new workloads
4. **Evidence Bundle** - Collect logs, events, and pod specs
5. **Ticket and Notify** - Create ticket and send notifications

### Safety Guardrails

- Never act on `kube-system` or `security-system` namespaces
- Respect `security.response=manual` pod labels
- Never isolate control plane nodes
- All actions are audited in `action_logs`

## RuntimeEvent Schema

```json
{
  "ts": "2026-01-10T21:12:33.123Z",
  "cluster_id": "kind-local",
  "node_id": "kind-worker",
  "event_type": "process_exec",
  "event_id": "uuid",
  "process": {
    "pid": 123,
    "ppid": 1,
    "uid": 0,
    "gid": 0,
    "exe": "/bin/bash",
    "cmdline": "bash -i",
    "cwd": "/",
    "has_tty": true,
    "capabilities_added": ["SYS_ADMIN"]
  },
  "container": {
    "container_id": "containerd://...",
    "image": "nginx:1.25",
    "image_digest": "sha256:...",
    "pod": "vuln-nginx-7c9b",
    "namespace": "prod",
    "service_account": "default",
    "labels": {"app": "vuln-nginx"}
  },
  "network": {
    "dst_ip": "10.0.0.12",
    "dst_port": 4444,
    "proto": "tcp",
    "dst_domain": ""
  },
  "raw_ref": "s3://bucket/raw/.../file.jsonl.gz#offset=12345"
}
```

## Testing

### Run Unit Tests

```bash
cd detect/matcher
go test -v ./...
```

### Run Golden Tests

```bash
go test -v ./... -run TestRuleEngine_FromFixture
```

### End-to-End Test

```bash
# Start kind cluster
kind create cluster --name kubeguard-test

# Deploy
helm install kubeguard ./deploy/helm/kubeguard -n security-system --create-namespace

# Run attacks
./test/attacks/run_all.sh

# Verify alerts
kubectl port-forward svc/kubeguard-incident 8081:8081 -n security-system &
curl http://localhost:8081/v1/alerts | jq
```

## Observability

### Metrics Exposed

- `kubeguard_ingest_events_total` - Total events ingested
- `kubeguard_detection_latency_seconds` - Detection latency histogram
- `kubeguard_alerts_total` - Alerts by rule and severity
- `kubeguard_response_actions_total` - Response actions by type and status

### Grafana Dashboards

Import dashboards from `deploy/grafana/` for:
- Event pipeline throughput
- Detection latency p95
- Alert rate by rule
- Response success rate

## Security

- **mTLS**: Between sensor and ingest, and between all internal services
- **JWT Auth**: For UI authentication
- **RBAC**: Least privilege Kubernetes permissions
- **No cluster-admin**: Response orchestrator uses minimal required permissions

## Development

### Build Services

```bash
# Ingest
cd ingest && go build -o bin/ingest .

# Detect
cd detect && go build -o bin/detect .

# Incident
cd incident && go build -o bin/incident .

# Respond
cd respond && go build -o bin/respond .
```

### Build UI

```bash
cd ui
npm install
npm run build
```

### Local Development

```bash
# Start dependencies
docker-compose up -d nats redis postgres minio

# Run services
go run ./ingest/main.go &
go run ./enrich/main.go &
go run ./detect/main.go &
go run ./incident/main.go &
go run ./respond/main.go &

# Run UI
cd ui && npm run dev
```

## License

MIT License - See [LICENSE](LICENSE) for details.

## Acknowledgments

- [Falco](https://falco.org/) - Cloud Native Runtime Security
- [CEL](https://github.com/google/cel-go) - Common Expression Language
- [NATS](https://nats.io/) - Cloud Native Messaging
