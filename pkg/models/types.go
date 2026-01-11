package models

import (
	"time"
)

type RuntimeEvent struct {
	Timestamp time.Time      `json:"ts"`
	ClusterID string         `json:"cluster_id"`
	NodeID    string         `json:"node_id"`
	EventType string         `json:"event_type"`
	EventID   string         `json:"event_id"`
	Process   *ProcessInfo   `json:"process,omitempty"`
	Container *ContainerInfo `json:"container,omitempty"`
	Network   *NetworkInfo   `json:"network,omitempty"`
	RawRef    string         `json:"raw_ref,omitempty"`
}

type ProcessInfo struct {
	PID               int      `json:"pid"`
	PPID              int      `json:"ppid"`
	UID               int      `json:"uid"`
	GID               int      `json:"gid"`
	Exe               string   `json:"exe"`
	Cmdline           string   `json:"cmdline"`
	Cwd               string   `json:"cwd"`
	HasTTY            bool     `json:"has_tty"`
	CapabilitiesAdded []string `json:"capabilities_added,omitempty"`
}

type ContainerInfo struct {
	ContainerID    string            `json:"container_id"`
	Image          string            `json:"image"`
	ImageDigest    string            `json:"image_digest"`
	Pod            string            `json:"pod"`
	Namespace      string            `json:"namespace"`
	ServiceAccount string            `json:"service_account"`
	Labels         map[string]string `json:"labels"`
}

type NetworkInfo struct {
	DstIP     string `json:"dst_ip"`
	DstPort   int    `json:"dst_port"`
	Proto     string `json:"proto"`
	DstDomain string `json:"dst_domain"`
}

type Alert struct {
	ID          string        `json:"id"`
	Timestamp   time.Time     `json:"timestamp"`
	RuleName    string        `json:"rule_name"`
	Severity    string        `json:"severity"`
	Description string        `json:"description"`
	Event       *RuntimeEvent `json:"event"`
	IncidentID  string        `json:"incident_id,omitempty"`
	Response    string        `json:"response,omitempty"` // Requested response action
}

type Incident struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"` // open, investigating, contained, resolved
	Severity        string    `json:"severity"`
	Title           string    `json:"title"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	AlertIDs        []string  `json:"alert_ids"`
	TriggeringEvent string    `json:"triggering_event_id"`
}

type Rule struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Severity    string `json:"severity" yaml:"severity"`
	Condition   string `json:"condition" yaml:"condition"` // CEL expression
	Response    string `json:"response" yaml:"response"`   // playbook name
	Enabled     bool   `json:"enabled" yaml:"enabled"`
}
