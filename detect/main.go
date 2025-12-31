package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/podwatch/podwatch/detect/matcher"
	"github.com/podwatch/podwatch/pkg/logging"
	"github.com/podwatch/podwatch/pkg/models"
)

var (
	natsConn *nats.Conn
	logger   *logging.Logger
)

// Rule to attack type mapping
var ruleAttackType = map[string]string{
	"rule-1": logging.AttackShellSpawn,
	"rule-2": logging.AttackTokenTheft,
	"rule-3": logging.AttackReverseShell,
	"rule-4": logging.AttackPrivilegeEscalation,
	"rule-5": logging.AttackPackageInstall,
}

func main() {
	logger = logging.NewLogger("podwatch-detect", "engine")

	// 1. Rules
	rules := []models.Rule{
		{
			ID:          "rule-1",
			Name:        "Shell Spawn in Prod",
			Description: "Bash or sh spawned in production namespace",
			Severity:    "high",
			Condition:   `event.process.exe in ['/bin/bash', '/bin/sh', '/usr/bin/bash'] && event.container.namespace == 'prod'`,
			Response:    "kill_pod",
			Enabled:     true,
		},
		{
			ID:          "rule-2",
			Name:        "Service Account Token Read",
			Description: "Process reading service account token",
			Severity:    "high",
			Condition:   `event.event_type == 'file_open' && event.process.cmdline.contains('/var/run/secrets/kubernetes.io/serviceaccount/token')`,
			Response:    "quarantine_namespace",
			Enabled:     true,
		},
		{
			ID:          "rule-3",
			Name:        "Reverse Shell",
			Description: "Network connection to external IP with shell process",
			Severity:    "critical",
			Condition:   `event.event_type == 'network_connect' && (event.process.exe.endsWith('bash') || event.process.exe.endsWith('sh')) && event.network.dst_ip != '' && !event.network.dst_ip.startsWith('10.') && !event.network.dst_ip.startsWith('192.168.') && !event.network.dst_ip.startsWith('172.')`,
			Response:    "kill_pod",
			Enabled:     true,
		},
		{
			ID:          "rule-4",
			Name:        "Privilege Escalation",
			Description: "Container added sensitive capabilities",
			Severity:    "critical",
			Condition:   `event.process.capabilities_added.exists(c, c == 'SYS_ADMIN' || c == 'NET_ADMIN')`,
			Response:    "isolate_node",
			Enabled:     true,
		},
		{
			ID:          "rule-5",
			Name:        "Package Manager in Prod",
			Description: "apt or apk executed in prod",
			Severity:    "medium",
			Condition:   `event.process.exe in ['/usr/bin/apt', '/sbin/apk', '/usr/bin/yum'] && event.container.namespace == 'prod'`,
			Response:    "",
			Enabled:     true,
		},
	}

	// 2. Engine
	engine, err := matcher.NewRuleEngine(rules)
	if err != nil {
		logger.Error("Failed to initialize rule engine", err, nil)
		os.Exit(1)
	}

	// 3. NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	for i := 0; i < 5; i++ {
		natsConn, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Error("Failed to connect to NATS", err, nil)
		os.Exit(1)
	}
	defer natsConn.Close()

	logger.Info("Detection engine started", map[string]interface{}{
		"rules_loaded": len(rules),
		"nats_url":     natsURL,
	})

	// 4. Subscribe
	handleEvent := func(msg *nats.Msg) {
		var event models.RuntimeEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Error("Failed to decode event", err, nil)
			return
		}

		// Evaluate
		alerts, err := engine.Evaluate(event)
		if err != nil {
			logger.Error("Rule evaluation failed", err, map[string]interface{}{
				"event_id": event.EventID,
			})
			return
		}

		for _, alert := range alerts {
			alert.ID = uuid.New().String()
			alert.Timestamp = time.Now().UTC()

			// Build target info from event
			target := buildTargetInfo(&event)

			// Get attack type and build attack context
			attackType := ruleAttackType[alert.RuleName]
			if attackType == "" {
				attackType = "unknown"
			}

			indicators := buildIndicators(&event)
			attackCtx := logging.GetAttackContext(attackType, alert.RuleName, alert.ID, indicators)
			attackCtx.Severity = alert.Severity

			// Log the attack detection
			logger.Attack(alert.Description, attackCtx, target)

			// Also log structured alert
			logger.Log(logging.SecurityEvent{
				Level:   logging.LevelAlert,
				Message: alert.Description,
				AlertID: alert.ID,
				EventID: event.EventID,
				Attack:  attackCtx,
				Target:  target,
				Metadata: map[string]interface{}{
					"response_action": alert.Response,
					"event_type":      event.EventType,
				},
			})

			data, _ := json.Marshal(alert)
			if err := natsConn.Publish("alerts", data); err != nil {
				logger.Error("Failed to publish alert", err, map[string]interface{}{
					"alert_id": alert.ID,
				})
			}
		}
	}

	// Subscribe to enriched events
	_, err = natsConn.QueueSubscribe("events.enriched", "detect-workers", handleEvent)
	if err != nil {
		logger.Error("Failed to subscribe to enriched events", err, nil)
		os.Exit(1)
	}

	// Also subscribe to raw events for demo without enrich service
	_, err = natsConn.Subscribe("events.raw.>", handleEvent)
	if err != nil {
		logger.Error("Failed to subscribe to raw events", err, nil)
		os.Exit(1)
	}

	logger.Info("Subscribed to event streams", map[string]interface{}{
		"streams": []string{"events.enriched", "events.raw.>"},
	})

	select {}
}

func buildTargetInfo(event *models.RuntimeEvent) *logging.TargetInfo {
	target := &logging.TargetInfo{
		ClusterID: event.ClusterID,
		Node:      event.NodeID,
	}

	if event.Container != nil {
		target.Namespace = event.Container.Namespace
		target.Pod = event.Container.Pod
		target.ContainerID = event.Container.ContainerID
		target.Image = event.Container.Image
		target.ImageDigest = event.Container.ImageDigest
		target.ServiceAccount = event.Container.ServiceAccount
	}

	if event.Process != nil {
		target.ProcessID = event.Process.PID
		target.ProcessPath = event.Process.Exe
		target.CommandLine = event.Process.Cmdline
		target.ParentPID = event.Process.PPID
		target.UID = event.Process.UID
	}

	if event.Network != nil {
		target.DestIP = event.Network.DstIP
		target.DestPort = event.Network.DstPort
		target.Protocol = event.Network.Proto
	}

	return target
}

func buildIndicators(event *models.RuntimeEvent) []string {
	var indicators []string

	if event.Process != nil {
		if event.Process.Exe != "" {
			indicators = append(indicators, "exe:"+event.Process.Exe)
		}
		if event.Process.Cmdline != "" {
			// Truncate long command lines
			cmd := event.Process.Cmdline
			if len(cmd) > 100 {
				cmd = cmd[:100] + "..."
			}
			indicators = append(indicators, "cmdline:"+cmd)
		}
		for _, cap := range event.Process.CapabilitiesAdded {
			indicators = append(indicators, "capability:"+cap)
		}
	}

	if event.Network != nil && event.Network.DstIP != "" {
		indicators = append(indicators, "dest_ip:"+event.Network.DstIP)
		if event.Network.DstPort > 0 {
			indicators = append(indicators, fmt.Sprintf("dest_port:%d", event.Network.DstPort))
		}
	}

	if event.Container != nil {
		indicators = append(indicators, "namespace:"+event.Container.Namespace)
		indicators = append(indicators, "image:"+event.Container.Image)
	}

	return indicators
}
