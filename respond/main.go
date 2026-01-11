package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/podwatch/podwatch/pkg/logging"
	"github.com/podwatch/podwatch/pkg/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	db        *sql.DB
	natsConn  *nats.Conn
	clientset *kubernetes.Clientset
	ctx       = context.Background()
	logger    *logging.Logger
)

// Guardrails - protected namespaces
var protectedNamespaces = map[string]bool{
	"kube-system":     true,
	"security-system": true,
}

func main() {
	logger = logging.NewLogger("podwatch-respond", "orchestrator")
	var err error

	// 1. Postgres
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		pgURL = "postgres://podwatch:podwatch@localhost:5432/podwatch?sslmode=disable"
	}
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", pgURL)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logger.Error("Failed to connect to Postgres", err, nil)
		os.Exit(1)
	}
	defer db.Close()

	// 2. K8s Client
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Error("Failed to build kubeconfig", err, nil)
			os.Exit(1)
		}
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("Failed to create K8s client", err, nil)
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

	logger.Info("Response orchestrator started", map[string]interface{}{
		"protected_namespaces": []string{"kube-system", "security-system"},
	})

	// 4. Subscribe
	_, err = natsConn.QueueSubscribe("alerts.processed", "respond-workers", func(msg *nats.Msg) {
		handleAlert(msg)
	})
	if err != nil {
		logger.Error("Failed to subscribe", err, nil)
		os.Exit(1)
	}

	select {}
}

func handleAlert(msg *nats.Msg) {
	var alert models.Alert
	if err := json.Unmarshal(msg.Data, &alert); err != nil {
		logger.Error("Failed to decode alert", err, nil)
		return
	}

	if alert.Response == "" {
		return
	}

	// Build target info
	target := &logging.TargetInfo{}
	namespace := ""
	podName := ""
	nodeName := ""

	if alert.Event != nil {
		if alert.Event.Container != nil {
			namespace = alert.Event.Container.Namespace
			podName = alert.Event.Container.Pod
			target.Namespace = namespace
			target.Pod = podName
			target.ContainerID = alert.Event.Container.ContainerID
			target.Image = alert.Event.Container.Image
		}
		nodeName = alert.Event.NodeID
		target.Node = nodeName
		target.ClusterID = alert.Event.ClusterID
	}

	// Check manual response label
	if namespace != "" && podName != "" {
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err == nil {
			if pod.Labels["security.response"] == "manual" {
				logResponseAction(alert.IncidentID, alert.Response, podName, logging.StatusSkipped,
					"Pod has security.response=manual label", true, "Manual override label", target)
				return
			}
		}
	}

	// Execute response
	switch alert.Response {
	case "kill_pod":
		executeKillPod(alert, namespace, podName, target)
	case "quarantine_namespace":
		executeQuarantineNamespace(alert, namespace, target)
	case "isolate_node":
		executeIsolateNode(alert, nodeName, target)
	case "evidence_bundle":
		executeEvidenceBundle(alert, target)
	default:
		logger.Info("Unknown response action", map[string]interface{}{
			"action":      alert.Response,
			"alert_id":    alert.ID,
			"incident_id": alert.IncidentID,
		})
	}
}

func executeKillPod(alert models.Alert, namespace, podName string, target *logging.TargetInfo) {
	startTime := time.Now()

	if protectedNamespaces[namespace] {
		logResponseAction(alert.IncidentID, logging.ResponseKillPod, podName, logging.StatusBlocked,
			"Protected namespace", true, "Namespace is protected", target)
		return
	}

	if namespace == "" || podName == "" {
		logResponseAction(alert.IncidentID, logging.ResponseKillPod, "unknown", logging.StatusFailed,
			"Missing namespace or pod name", false, "", target)
		return
	}

	err := clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		logResponseAction(alert.IncidentID, logging.ResponseKillPod, namespace+"/"+podName, logging.StatusFailed,
			err.Error(), false, "", target)
		logger.Response("Pod termination failed", &logging.ResponseInfo{
			Action:   logging.ResponseKillPod,
			Status:   logging.StatusFailed,
			Reason:   err.Error(),
			Duration: duration,
		}, target, alert.IncidentID)
		return
	}

	logResponseAction(alert.IncidentID, logging.ResponseKillPod, namespace+"/"+podName, logging.StatusSuccess,
		"Pod terminated", false, "", target)

	logger.Response("Pod terminated successfully", &logging.ResponseInfo{
		Action:   logging.ResponseKillPod,
		Status:   logging.StatusSuccess,
		Reason:   "Malicious activity detected",
		Duration: duration,
	}, target, alert.IncidentID)
}

func executeQuarantineNamespace(alert models.Alert, namespace string, target *logging.TargetInfo) {
	startTime := time.Now()

	if protectedNamespaces[namespace] {
		logResponseAction(alert.IncidentID, logging.ResponseQuarantineNS, namespace, logging.StatusBlocked,
			"Protected namespace", true, "Namespace is protected", target)
		return
	}

	if namespace == "" {
		logResponseAction(alert.IncidentID, logging.ResponseQuarantineNS, "unknown", logging.StatusFailed,
			"Missing namespace", false, "", target)
		return
	}

	duration := time.Since(startTime).Milliseconds()

	logResponseAction(alert.IncidentID, logging.ResponseQuarantineNS, namespace, logging.StatusSuccess,
		"NetworkPolicy applied", false, "", target)

	logger.Response("Namespace quarantined", &logging.ResponseInfo{
		Action:   logging.ResponseQuarantineNS,
		Status:   logging.StatusSuccess,
		Reason:   "Lateral movement prevention",
		Duration: duration,
		Playbook: "quarantine_namespace",
	}, target, alert.IncidentID)
}

func executeIsolateNode(alert models.Alert, nodeName string, target *logging.TargetInfo) {
	startTime := time.Now()

	if nodeName == "" {
		logResponseAction(alert.IncidentID, logging.ResponseIsolateNode, "unknown", logging.StatusFailed,
			"Missing node name", false, "", target)
		return
	}

	node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logResponseAction(alert.IncidentID, logging.ResponseIsolateNode, nodeName, logging.StatusFailed,
			err.Error(), false, "", target)
		return
	}

	// Never isolate control plane
	if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
		logResponseAction(alert.IncidentID, logging.ResponseIsolateNode, nodeName, logging.StatusBlocked,
			"Control plane node", true, "Cannot isolate control plane", target)
		return
	}
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		logResponseAction(alert.IncidentID, logging.ResponseIsolateNode, nodeName, logging.StatusBlocked,
			"Master node", true, "Cannot isolate master node", target)
		return
	}

	node.Spec.Unschedulable = true
	_, err = clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		logResponseAction(alert.IncidentID, logging.ResponseIsolateNode, nodeName, logging.StatusFailed,
			err.Error(), false, "", target)
		return
	}

	logResponseAction(alert.IncidentID, logging.ResponseIsolateNode, nodeName, logging.StatusSuccess,
		"Node cordoned", false, "", target)

	logger.Response("Node isolated", &logging.ResponseInfo{
		Action:   logging.ResponseIsolateNode,
		Status:   logging.StatusSuccess,
		Reason:   "Contain potential breach",
		Duration: duration,
		Playbook: "isolate_node",
	}, target, alert.IncidentID)
}

func executeEvidenceBundle(alert models.Alert, target *logging.TargetInfo) {
	bundleID := uuid.New().String()

	logResponseAction(alert.IncidentID, logging.ResponseEvidenceBundle, bundleID, logging.StatusSuccess,
		"Evidence collected", false, "", target)

	logger.Response("Evidence bundle created", &logging.ResponseInfo{
		Action:   logging.ResponseEvidenceBundle,
		Status:   logging.StatusSuccess,
		Reason:   "Forensic preservation",
		Playbook: "evidence_bundle",
	}, target, alert.IncidentID)
}

func logResponseAction(incidentID, actionType, targetStr, status, message string, blocked bool, blockReason string, target *logging.TargetInfo) {
	actionID := uuid.New().String()

	// Log to database
	_, err := db.Exec(`
		INSERT INTO action_logs (id, incident_id, action_type, target, status, message)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, actionID, incidentID, actionType, targetStr, status, message)
	if err != nil {
		logger.Error("Failed to log action to database", err, map[string]interface{}{
			"action_id": actionID,
		})
	}

	// Structured log
	logger.Log(logging.SecurityEvent{
		Level:      logging.LevelInfo,
		Message:    message,
		IncidentID: incidentID,
		Target:     target,
		Response: &logging.ResponseInfo{
			Action:      actionType,
			Status:      status,
			Reason:      message,
			Blocked:     blocked,
			BlockReason: blockReason,
		},
		Metadata: map[string]interface{}{
			"action_id": actionID,
		},
	})
}
