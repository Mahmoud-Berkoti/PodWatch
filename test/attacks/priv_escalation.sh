#!/bin/bash
# Privilege Escalation Simulation
# This script simulates a privilege escalation attempt

set -e

echo "=================================================="
echo "KubeGuard Attack Simulation: Privilege Escalation"
echo "=================================================="

NAMESPACE="${NAMESPACE:-attacker-lab}"
POD_NAME="${POD_NAME:-priv-esc}"

echo ""
echo "[*] Creating privileged pod in namespace: $NAMESPACE"

kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Create a pod with elevated privileges
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: $POD_NAME
  namespace: $NAMESPACE
  labels:
    app: priv-esc
    env: lab
spec:
  containers:
  - name: privileged
    image: alpine:latest
    command: ["sleep", "3600"]
    securityContext:
      privileged: true
      capabilities:
        add:
        - SYS_ADMIN
        - NET_ADMIN
        - SYS_PTRACE
EOF

echo "[*] Waiting for pod to be ready..."
kubectl wait --for=condition=Ready pod/$POD_NAME -n $NAMESPACE --timeout=60s

echo ""
echo "[!] Pod created with elevated capabilities..."
echo "[!] This should trigger: Privilege Escalation rule"
echo ""

# The pod creation itself with SYS_ADMIN capability should trigger
# But let's also try to use the capability
kubectl exec -n $NAMESPACE $POD_NAME -- sh -c "
  echo 'Checking capabilities...'
  cat /proc/self/status | grep Cap
  echo ''
  echo 'Attempting mount (requires SYS_ADMIN)...'
  mount -t tmpfs none /mnt 2>/dev/null && echo 'Mount successful!' || echo 'Mount failed'
"

echo ""
echo "[*] Attack simulation completed. Check KubeGuard for alerts."
echo ""
echo "Expected:"
echo "  - Alert: Privilege Escalation"
echo "  - Severity: critical"
echo "  - Response: isolate_node"
echo ""

echo "[*] To cleanup: kubectl delete pod $POD_NAME -n $NAMESPACE"
