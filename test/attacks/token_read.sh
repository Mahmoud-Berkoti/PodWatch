#!/bin/bash
# Service Account Token Read Attack Simulation
# This script simulates reading the SA token inside a container

set -e

echo "================================================"
echo "KubeGuard Attack Simulation: Token Read"
echo "================================================"

NAMESPACE="${NAMESPACE:-prod}"
POD_NAME="${POD_NAME:-token-reader}"

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
    app: token-reader
    env: prod
spec:
  containers:
  - name: attacker
    image: busybox:latest
    command: ["sleep", "3600"]
EOF

echo "[*] Waiting for pod to be ready..."
kubectl wait --for=condition=Ready pod/$POD_NAME -n $NAMESPACE --timeout=60s

echo ""
echo "[!] Executing token read attack..."
echo "[!] This should trigger: Service Account Token Read rule"
echo ""

# Read the service account token - this should trigger detection
kubectl exec -n $NAMESPACE $POD_NAME -- cat /var/run/secrets/kubernetes.io/serviceaccount/token || true

echo ""
echo "[*] Attack executed. Check KubeGuard for alerts."
echo ""
echo "Expected:"
echo "  - Alert: Service Account Token Read"
echo "  - Severity: high"
echo "  - Response: quarantine_namespace"
echo ""

echo "[*] Pod $POD_NAME is still running for inspection."
echo "[*] To cleanup: kubectl delete pod $POD_NAME -n $NAMESPACE"
