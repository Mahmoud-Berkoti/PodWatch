package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LevelDebug    LogLevel = "DEBUG"
	LevelInfo     LogLevel = "INFO"
	LevelWarn     LogLevel = "WARN"
	LevelError    LogLevel = "ERROR"
	LevelCritical LogLevel = "CRITICAL"
	LevelAlert    LogLevel = "ALERT"  // Security alert
	LevelBreach   LogLevel = "BREACH" // Confirmed breach
	LevelAttack   LogLevel = "ATTACK" // Attack detected
)

// SecurityEvent represents a structured security log entry
type SecurityEvent struct {
	// Core fields
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Service   string   `json:"service"`
	Component string   `json:"component"`
	Message   string   `json:"message"`

	// Correlation
	TraceID    string `json:"trace_id,omitempty"`
	SpanID     string `json:"span_id,omitempty"`
	EventID    string `json:"event_id,omitempty"`
	IncidentID string `json:"incident_id,omitempty"`
	AlertID    string `json:"alert_id,omitempty"`

	// Attack Context
	Attack *AttackContext `json:"attack,omitempty"`

	// Target Information
	Target *TargetInfo `json:"target,omitempty"`

	// Response Actions
	Response *ResponseInfo `json:"response,omitempty"`

	// Additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AttackContext provides detailed attack information
type AttackContext struct {
	Type           string   `json:"type"`       // e.g., "shell_spawn", "token_theft", "privilege_escalation"
	Technique      string   `json:"technique"`  // MITRE ATT&CK technique ID
	TacticID       string   `json:"tactic_id"`  // MITRE ATT&CK tactic
	Severity       string   `json:"severity"`   // critical, high, medium, low
	Confidence     float64  `json:"confidence"` // 0.0-1.0 confidence score
	RuleName       string   `json:"rule_name"`  // Detection rule that triggered
	RuleID         string   `json:"rule_id"`
	Indicators     []string `json:"indicators"`       // IOCs observed
	KillChainPhase string   `json:"kill_chain_phase"` // reconnaissance, weaponization, delivery, exploitation, installation, c2, actions
}

// TargetInfo identifies what was targeted
type TargetInfo struct {
	// Kubernetes context
	ClusterID      string `json:"cluster_id"`
	Namespace      string `json:"namespace"`
	Pod            string `json:"pod"`
	Container      string `json:"container"`
	ContainerID    string `json:"container_id"`
	Node           string `json:"node"`
	ServiceAccount string `json:"service_account"`

	// Image information
	Image       string `json:"image"`
	ImageDigest string `json:"image_digest"`

	// Process context
	ProcessID   int    `json:"process_id,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
	ProcessPath string `json:"process_path,omitempty"`
	CommandLine string `json:"command_line,omitempty"`
	ParentPID   int    `json:"parent_pid,omitempty"`
	User        string `json:"user,omitempty"`
	UID         int    `json:"uid,omitempty"`

	// Network context
	SourceIP   string `json:"source_ip,omitempty"`
	SourcePort int    `json:"source_port,omitempty"`
	DestIP     string `json:"dest_ip,omitempty"`
	DestPort   int    `json:"dest_port,omitempty"`
	Protocol   string `json:"protocol,omitempty"`

	// File context
	FilePath      string `json:"file_path,omitempty"`
	FileOperation string `json:"file_operation,omitempty"`
}

// ResponseInfo tracks automated response actions
type ResponseInfo struct {
	Action      string `json:"action"`  // kill_pod, quarantine_namespace, isolate_node
	Status      string `json:"status"`  // pending, executing, success, failed, blocked
	Reason      string `json:"reason"`  // Why this action was taken
	Blocked     bool   `json:"blocked"` // Was action blocked by guardrails
	BlockReason string `json:"block_reason,omitempty"`
	Duration    int64  `json:"duration_ms"` // Execution time in milliseconds
	Playbook    string `json:"playbook"`    // Playbook that triggered this
}

// Logger provides structured security logging
type Logger struct {
	service   string
	component string
	output    *json.Encoder
}

// NewLogger creates a new structured logger
func NewLogger(service, component string) *Logger {
	return &Logger{
		service:   service,
		component: component,
		output:    json.NewEncoder(os.Stdout),
	}
}

// Log writes a structured log entry
func (l *Logger) Log(event SecurityEvent) {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	event.Service = l.service
	event.Component = l.component
	l.output.Encode(event)
}

// Attack logs an attack detection event
func (l *Logger) Attack(msg string, attack *AttackContext, target *TargetInfo) {
	l.Log(SecurityEvent{
		Level:   LevelAttack,
		Message: msg,
		Attack:  attack,
		Target:  target,
	})
}

// Breach logs a confirmed breach
func (l *Logger) Breach(msg string, attack *AttackContext, target *TargetInfo, incidentID string) {
	l.Log(SecurityEvent{
		Level:      LevelBreach,
		Message:    msg,
		Attack:     attack,
		Target:     target,
		IncidentID: incidentID,
	})
}

// Alert logs a security alert
func (l *Logger) Alert(msg string, alertID string, attack *AttackContext, target *TargetInfo) {
	l.Log(SecurityEvent{
		Level:   LevelAlert,
		Message: msg,
		AlertID: alertID,
		Attack:  attack,
		Target:  target,
	})
}

// Response logs a response action
func (l *Logger) Response(msg string, response *ResponseInfo, target *TargetInfo, incidentID string) {
	l.Log(SecurityEvent{
		Level:      LevelInfo,
		Message:    msg,
		Response:   response,
		Target:     target,
		IncidentID: incidentID,
	})
}

// Info logs an informational message
func (l *Logger) Info(msg string, metadata map[string]interface{}) {
	l.Log(SecurityEvent{
		Level:    LevelInfo,
		Message:  msg,
		Metadata: metadata,
	})
}

// Error logs an error
func (l *Logger) Error(msg string, err error, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["error"] = err.Error()
	l.Log(SecurityEvent{
		Level:    LevelError,
		Message:  msg,
		Metadata: metadata,
	})
}

// WithTrace returns a new logger with trace context
func (l *Logger) WithTrace(traceID, spanID string) *Logger {
	// In a real implementation, this would create a child logger with context
	return l
}

// FormatForSearch returns a search-friendly string representation
func (e *SecurityEvent) FormatForSearch() string {
	parts := []string{
		fmt.Sprintf("[%s]", e.Level),
		e.Timestamp,
		e.Service,
	}

	if e.Attack != nil {
		parts = append(parts, fmt.Sprintf("attack_type=%s", e.Attack.Type))
		parts = append(parts, fmt.Sprintf("severity=%s", e.Attack.Severity))
		parts = append(parts, fmt.Sprintf("rule=%s", e.Attack.RuleName))
	}

	if e.Target != nil {
		if e.Target.Namespace != "" {
			parts = append(parts, fmt.Sprintf("namespace=%s", e.Target.Namespace))
		}
		if e.Target.Pod != "" {
			parts = append(parts, fmt.Sprintf("pod=%s", e.Target.Pod))
		}
		if e.Target.ProcessPath != "" {
			parts = append(parts, fmt.Sprintf("process=%s", e.Target.ProcessPath))
		}
	}

	if e.IncidentID != "" {
		parts = append(parts, fmt.Sprintf("incident=%s", e.IncidentID))
	}

	parts = append(parts, e.Message)

	return fmt.Sprintf("%s", parts)
}
