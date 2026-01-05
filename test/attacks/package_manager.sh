#!/bin/bash
# Package Manager Attack Simulation
# This script simulates running a package manager in production

set -e

echo "================================================"
echo "KubeGuard Attack Simulation: Package Manager"
echo "================================================"

NAMESPACE="${NAMESPACE:-prod}"
POD_NAME="${POD_NAME:-pkg-install}"

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
    app: pkg-install
    env: prod
spec:
  containers:
  - name: ubuntu
    image: ubuntu:22.04
    command: ["sleep", "3600"]
EOF

echo "[*] Waiting for pod to be ready..."
kubectl wait --for=condition=Ready pod/$POD_NAME -n $NAMESPACE --timeout=60s

echo ""
echo "[!] Executing package manager..."
echo "[!] This should trigger: Package Manager in Production rule"
echo ""

# Try to run apt update - this triggers detection even if it fails
kubectl exec -n $NAMESPACE $POD_NAME -- bash -c "
  echo 'Running apt update...'
  apt update 2>&1 | head -5 || true
  echo 'Package manager executed!'
"

echo ""
echo "[*] Attack simulation completed. Check KubeGuard for alerts."
echo ""
echo "Expected:"
echo "  - Alert: Package Manager in Production"
echo "  - Severity: medium"
echo "  - Response: (alert only)"
echo ""

echo "[*] To cleanup: kubectl delete pod $POD_NAME -n $NAMESPACE"
