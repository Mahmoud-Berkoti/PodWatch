#!/bin/bash
# Run all attack simulations
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "============================================"
echo "KubeGuard Full Attack Simulation Suite"
echo "============================================"
echo ""

# Create test namespaces
echo "[*] Setting up test namespaces..."
kubectl create namespace prod --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace attacker-lab --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "============================================"
echo "Running Attack 1: Shell Spawn"
echo "============================================"
NAMESPACE=prod POD_NAME=vuln-nginx-1 bash "$SCRIPT_DIR/shell_spawn.sh"
sleep 5

echo ""
echo "============================================"
echo "Running Attack 2: Token Read"
echo "============================================"
NAMESPACE=prod POD_NAME=token-reader-1 bash "$SCRIPT_DIR/token_read.sh"
sleep 5

echo ""
echo "============================================"
echo "Running Attack 3: Reverse Shell"
echo "============================================"
NAMESPACE=attacker-lab POD_NAME=reverse-shell-1 bash "$SCRIPT_DIR/reverse_shell.sh"
sleep 5

echo ""
echo "============================================"
echo "Running Attack 4: Privilege Escalation"
echo "============================================"
NAMESPACE=attacker-lab POD_NAME=priv-esc-1 bash "$SCRIPT_DIR/priv_escalation.sh"
sleep 5

echo ""
echo "============================================"
echo "Running Attack 5: Package Manager"
echo "============================================"
NAMESPACE=prod POD_NAME=pkg-install-1 bash "$SCRIPT_DIR/package_manager.sh"
sleep 5

echo ""
echo "============================================"
echo "Attack Simulation Complete!"
echo "============================================"
echo ""
echo "Check KubeGuard UI for:"
echo "  - 5+ Alerts generated"
echo "  - Multiple incidents created"
echo "  - Response actions executed"
echo ""
echo "To view alerts:"
echo "  kubectl port-forward svc/kubeguard-incident 8081:8081 -n security-system"
echo "  curl http://localhost:8081/v1/alerts"
echo ""
echo "To cleanup test pods:"
echo "  kubectl delete pods --all -n prod"
echo "  kubectl delete pods --all -n attacker-lab"
