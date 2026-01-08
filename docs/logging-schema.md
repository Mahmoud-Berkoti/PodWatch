# PodWatch Logging Schema

PodWatch uses a structured JSON logging format designed to make it easy to identify, search, and investigate security incidents.

## Log Levels

| Level | Description |
|-------|-------------|
| `DEBUG` | Detailed debugging information |
| `INFO` | General operational information |
| `WARN` | Warning conditions |
| `ERROR` | Error conditions |
| `CRITICAL` | Critical conditions requiring immediate attention |
| `ALERT` | Security alert generated |
| `ATTACK` | Attack pattern detected |
| `BREACH` | Confirmed security breach |

## Log Structure

Every log entry follows this JSON schema:

```json
{
  "timestamp": "2026-01-10T21:12:33.123456789Z",
  "level": "ATTACK",
  "service": "podwatch-detect",
  "component": "engine",
  "message": "Shell spawn detected in production namespace",
  
  "trace_id": "abc123",
  "span_id": "def456",
  "event_id": "evt-001",
  "incident_id": "inc-001",
  "alert_id": "alert-001",
  
  "attack": {
    "type": "shell_spawn",
    "technique": "T1609",
    "tactic_id": "TA0002",
    "severity": "high",
    "confidence": 0.95,
    "rule_name": "Shell Spawn in Prod",
    "rule_id": "rule-1",
    "indicators": [
      "exe:/bin/bash",
      "cmdline:bash -i",
      "namespace:prod"
    ],
    "kill_chain_phase": "exploitation"
  },
  
  "target": {
    "cluster_id": "prod-cluster",
    "namespace": "prod",
    "pod": "vuln-nginx-7c9b",
    "container": "nginx",
    "container_id": "containerd://abc123",
    "node": "worker-1",
    "service_account": "default",
    "image": "nginx:1.25",
    "image_digest": "sha256:abc123",
    "process_id": 12345,
    "process_name": "bash",
    "process_path": "/bin/bash",
    "command_line": "bash -i",
    "parent_pid": 1,
    "user": "root",
    "uid": 0,
    "dest_ip": "10.0.0.1",
    "dest_port": 4444,
    "protocol": "tcp"
  },
  
  "response": {
    "action": "kill_pod",
    "status": "success",
    "reason": "Malicious shell detected",
    "blocked": false,
    "block_reason": "",
    "duration_ms": 150,
    "playbook": "kill_pod"
  },
  
  "metadata": {
    "custom_field": "value"
  }
}
```

## Attack Types

| Type | Description | MITRE Technique |
|------|-------------|-----------------|
| `shell_spawn` | Interactive shell spawned in container | T1609 |
| `token_theft` | Service account token accessed | T1552.007 |
| `privilege_escalation` | Container gained elevated privileges | T1611 |
| `reverse_shell` | Outbound shell connection detected | T1609 |
| `lateral_movement` | Movement between pods/nodes | T1021 |
| `data_exfiltration` | Data leaving the cluster | T1041 |
| `crypto_mining` | Cryptocurrency mining detected | T1496 |
| `container_escape` | Attempt to escape container | T1611 |
| `credential_access` | Credential theft attempt | T1552 |
| `package_install` | Package manager execution | T1609 |

## MITRE ATT&CK Mapping

### Tactics

| ID | Name |
|----|------|
| TA0001 | Initial Access |
| TA0002 | Execution |
| TA0003 | Persistence |
| TA0004 | Privilege Escalation |
| TA0005 | Defense Evasion |
| TA0006 | Credential Access |
| TA0007 | Discovery |
| TA0008 | Lateral Movement |
| TA0009 | Collection |
| TA0010 | Exfiltration |
| TA0040 | Impact |

### Kill Chain Phases

| Phase | Description |
|-------|-------------|
| `reconnaissance` | Attacker gathering information |
| `weaponization` | Creating attack payload |
| `delivery` | Delivering payload to target |
| `exploitation` | Exploiting vulnerability |
| `installation` | Installing malware/backdoor |
| `command_and_control` | Establishing C2 channel |
| `actions_on_objectives` | Achieving attack goals |

## Searching Logs

### By Attack Type
```bash
grep '"type":"shell_spawn"' /var/log/podwatch/*.log
```

### By Severity
```bash
grep '"severity":"critical"' /var/log/podwatch/*.log
```

### By Namespace
```bash
grep '"namespace":"prod"' /var/log/podwatch/*.log
```

### By Incident
```bash
grep '"incident_id":"inc-12345"' /var/log/podwatch/*.log
```

### By Time Range (with jq)
```bash
cat /var/log/podwatch/*.log | jq 'select(.timestamp >= "2026-01-10T00:00:00Z" and .timestamp <= "2026-01-10T23:59:59Z")'
```

## OpenSearch/Elasticsearch Queries

### Find all critical attacks
```json
{
  "query": {
    "bool": {
      "must": [
        { "term": { "level": "ATTACK" } },
        { "term": { "attack.severity": "critical" } }
      ]
    }
  }
}
```

### Find attacks by technique
```json
{
  "query": {
    "term": { "attack.technique": "T1609" }
  }
}
```

### Find all actions for an incident
```json
{
  "query": {
    "term": { "incident_id": "inc-12345" }
  },
  "sort": [
    { "timestamp": "asc" }
  ]
}
```

## Response Status Values

| Status | Description |
|--------|-------------|
| `pending` | Action queued |
| `executing` | Action in progress |
| `success` | Action completed successfully |
| `failed` | Action failed |
| `blocked` | Action blocked by guardrails |
| `skipped` | Action skipped (manual override) |

## Example Log Entries

### Attack Detection
```json
{"timestamp":"2026-01-10T21:12:33.123Z","level":"ATTACK","service":"podwatch-detect","component":"engine","message":"Reverse shell detected","attack":{"type":"reverse_shell","technique":"T1609","tactic_id":"TA0002","severity":"critical","confidence":0.98,"rule_name":"Reverse Shell","indicators":["exe:/bin/bash","dest_ip:203.0.113.50","dest_port:4444"],"kill_chain_phase":"command_and_control"},"target":{"namespace":"prod","pod":"web-app-xyz","process_path":"/bin/bash","dest_ip":"203.0.113.50","dest_port":4444}}
```

### Response Action
```json
{"timestamp":"2026-01-10T21:12:33.456Z","level":"INFO","service":"podwatch-respond","component":"orchestrator","message":"Pod terminated successfully","incident_id":"inc-001","target":{"namespace":"prod","pod":"web-app-xyz"},"response":{"action":"kill_pod","status":"success","reason":"Malicious activity detected","duration_ms":125}}
```

### Blocked Action
```json
{"timestamp":"2026-01-10T21:12:33.789Z","level":"INFO","service":"podwatch-respond","component":"orchestrator","message":"Action blocked by guardrails","incident_id":"inc-002","target":{"namespace":"kube-system","pod":"coredns-abc"},"response":{"action":"kill_pod","status":"blocked","blocked":true,"block_reason":"Namespace is protected"}}
```

## Integration with SIEM

The structured format enables easy integration with:

- **Splunk**: Parse as JSON, create dashboards by attack type
- **Elastic/OpenSearch**: Index directly, use Kibana for visualization
- **Grafana Loki**: Label by level, service, attack type
- **Datadog**: Forward logs with proper tagging
- **AWS CloudWatch**: Filter by JSON fields

## Metrics Derived from Logs

Track these metrics from log data:

- Attacks per hour by type
- Mean time to detect (MTTD)
- Mean time to respond (MTTR)
- Response success rate by action type
- Blocked actions rate (guardrail effectiveness)
- Attacks by namespace/cluster
- Top triggered rules
