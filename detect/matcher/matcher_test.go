package matcher

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/podwatch/podwatch/pkg/models"
)

func TestRuleEngine_ShellSpawn(t *testing.T) {
	rules := []models.Rule{
		{
			ID:          "rule-shell-spawn",
			Name:        "Shell Spawn in Prod",
			Description: "Shell spawned in production",
			Severity:    "high",
			Condition:   `event.process.exe in ['/bin/bash', '/bin/sh'] && event.container.namespace == 'prod'`,
			Response:    "kill_pod",
			Enabled:     true,
		},
	}

	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test event that should match
	event := models.RuntimeEvent{
		Timestamp: time.Now(),
		ClusterID: "kind-local",
		NodeID:    "kind-worker",
		EventType: "process_exec",
		EventID:   "test-001",
		Process: &models.ProcessInfo{
			PID:     12345,
			Exe:     "/bin/bash",
			Cmdline: "bash -i",
		},
		Container: &models.ContainerInfo{
			ContainerID: "containerd://abc123",
			Namespace:   "prod",
			Pod:         "test-pod",
		},
	}

	alerts, err := engine.Evaluate(event)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].RuleName != "Shell Spawn in Prod" {
		t.Errorf("Expected rule name 'Shell Spawn in Prod', got '%s'", alerts[0].RuleName)
	}

	if alerts[0].Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", alerts[0].Severity)
	}
}

func TestRuleEngine_NoMatch(t *testing.T) {
	rules := []models.Rule{
		{
			ID:        "rule-shell-spawn",
			Name:      "Shell Spawn in Prod",
			Severity:  "high",
			Condition: `event.process.exe in ['/bin/bash'] && event.container.namespace == 'prod'`,
			Enabled:   true,
		},
	}

	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Event in staging namespace should not match
	event := models.RuntimeEvent{
		Timestamp: time.Now(),
		EventType: "process_exec",
		Process: &models.ProcessInfo{
			Exe: "/bin/bash",
		},
		Container: &models.ContainerInfo{
			Namespace: "staging", // Not prod
		},
	}

	alerts, err := engine.Evaluate(event)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(alerts) != 0 {
		t.Fatalf("Expected 0 alerts, got %d", len(alerts))
	}
}

func TestRuleEngine_PrivilegeEscalation(t *testing.T) {
	rules := []models.Rule{
		{
			ID:        "rule-priv-esc",
			Name:      "Privilege Escalation",
			Severity:  "critical",
			Condition: `event.process.capabilities_added.exists(c, c == 'SYS_ADMIN')`,
			Response:  "isolate_node",
			Enabled:   true,
		},
	}

	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	event := models.RuntimeEvent{
		Timestamp: time.Now(),
		EventType: "capability_change",
		Process: &models.ProcessInfo{
			Exe:               "/bin/mount",
			CapabilitiesAdded: []string{"SYS_ADMIN", "NET_ADMIN"},
		},
		Container: &models.ContainerInfo{
			Namespace: "attacker-lab",
		},
	}

	alerts, err := engine.Evaluate(event)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	if alerts[0].Response != "isolate_node" {
		t.Errorf("Expected response 'isolate_node', got '%s'", alerts[0].Response)
	}
}

func TestRuleEngine_FromFixture(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "shell_spawn_event.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("Fixture not found: %v", err)
	}

	var event models.RuntimeEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("Failed to unmarshal fixture: %v", err)
	}

	rules := []models.Rule{
		{
			ID:        "rule-shell-spawn",
			Name:      "Shell Spawn in Prod",
			Severity:  "high",
			Condition: `event.process.exe == '/bin/bash' && event.container.namespace == 'prod'`,
			Enabled:   true,
		},
	}

	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	alerts, err := engine.Evaluate(event)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert from fixture, got %d", len(alerts))
	}
}

func BenchmarkRuleEvaluation(b *testing.B) {
	rules := []models.Rule{
		{ID: "r1", Name: "Rule 1", Condition: `event.process.exe == '/bin/bash'`, Enabled: true},
		{ID: "r2", Name: "Rule 2", Condition: `event.container.namespace == 'prod'`, Enabled: true},
		{ID: "r3", Name: "Rule 3", Condition: `event.process.uid == 0`, Enabled: true},
	}

	engine, _ := NewRuleEngine(rules)

	event := models.RuntimeEvent{
		Timestamp: time.Now(),
		Process: &models.ProcessInfo{
			Exe: "/bin/bash",
			UID: 0,
		},
		Container: &models.ContainerInfo{
			Namespace: "prod",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Evaluate(event)
	}
}
