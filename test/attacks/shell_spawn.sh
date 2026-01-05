#!/bin/bash
# Shell Spawn Attack Simulation
# This script simulates a shell spawn attack in a container

set -e

echo "=========================================="
echo "KubeGuard Attack Simulation: Shell Spawn"
echo "=========================================="

NAMESPACE="${NAMESPACE:-prod}"
POD_NAME="${POD_NAME:-vuln-nginx}"

echo ""
echo "[*] Creating vulnerable pod in namespace: $NAMESPACE"

kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: $POD_NAME
  namespace: $NAMESPACE
  labels:
    app: vuln-nginx
    env: prod
spec:
  containers:
  - name: nginx
    image: nginx:1.25
    command: ["sleep", "3600"]
EOF

echo "[*] Waiting for pod to be ready..."
kubectl wait --for=condition=Ready pod/$POD_NAME -n $NAMESPACE --timeout=60s

echo ""
echo "[!] Executing shell spawn attack..."
echo "[!] This should trigger: Shell Spawn in Production rule"
echo ""

# Execute bash in the container - this should trigger the detection
kubectl exec -n $NAMESPACE $POD_NAME -- /bin/bash -c "echo 'Shell spawned successfully!'"

echo ""
echo "[*] Attack executed. Check KubeGuard for alerts."
echo ""
echo "Expected:"
echo "  - Alert: Shell Spawn in Production"
echo "  - Severity: high"
echo "  - Response: kill_pod"
echo ""

# Keep pod for inspection
echo "[*] Pod $POD_NAME is still running for inspection."
echo "[*] To cleanup: kubectl delete pod $POD_NAME -n $NAMESPACE"
