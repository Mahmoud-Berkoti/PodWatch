#!/bin/bash
# End-to-end test script for KubeGuard
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "============================================"
echo "KubeGuard End-to-End Test Suite"
echo "============================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

CLUSTER_NAME="podwatch-e2e"
NAMESPACE="security-system"

cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    kind delete cluster --name $CLUSTER_NAME 2>/dev/null || true
}

trap cleanup EXIT

# Step 1: Create Kind cluster
echo -e "\n${YELLOW}Step 1: Creating Kind cluster...${NC}"
kind create cluster --name $CLUSTER_NAME --wait 120s

# Step 2: Deploy KubeGuard
echo -e "\n${YELLOW}Step 2: Deploying KubeGuard...${NC}"
helm install podwatch "$PROJECT_ROOT/deploy/helm/podwatch" \
    -n $NAMESPACE --create-namespace \
    --set postgres.persistence.enabled=false \
    --set minio.persistence.enabled=false \
    --wait --timeout 5m

# Step 3: Wait for all pods
echo -e "\n${YELLOW}Step 3: Waiting for pods to be ready...${NC}"
kubectl wait --for=condition=ready pod --all -n $NAMESPACE --timeout=300s

# Step 4: Create test namespaces
echo -e "\n${YELLOW}Step 4: Creating test namespaces...${NC}"
kubectl create namespace prod --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace staging --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace attacker-lab --dry-run=client -o yaml | kubectl apply -f -

# Step 5: Run attack simulations
echo -e "\n${YELLOW}Step 5: Running attack simulations...${NC}"

# Deploy test workloads
echo "Deploying vulnerable workloads..."
kubectl run vuln-nginx --image=nginx:1.25 -n prod --command -- sleep 3600
kubectl run demo-api --image=busybox -n staging --command -- sleep 3600
kubectl run busybox-attacker --image=busybox -n attacker-lab --command -- sleep 3600

kubectl wait --for=condition=ready pod/vuln-nginx -n prod --timeout=60s
kubectl wait --for=condition=ready pod/demo-api -n staging --timeout=60s
kubectl wait --for=condition=ready pod/busybox-attacker -n attacker-lab --timeout=60s

# Execute attacks
echo "Executing shell spawn attack..."
kubectl exec -n prod vuln-nginx -- /bin/bash -c "echo 'Shell spawned!'" || true

echo "Executing token read attack..."
kubectl exec -n prod vuln-nginx -- cat /var/run/secrets/kubernetes.io/serviceaccount/token || true

echo "Executing package manager attack..."
kubectl exec -n staging demo-api -- sh -c "which apk && apk --version" || true

# Step 6: Wait for detection
echo -e "\n${YELLOW}Step 6: Waiting for detections (30s)...${NC}"
sleep 30

# Step 7: Verify alerts
echo -e "\n${YELLOW}Step 7: Verifying alerts...${NC}"
kubectl port-forward svc/podwatch-incident 8081:8081 -n $NAMESPACE &
PF_PID=$!
sleep 5

ALERTS=$(curl -s http://localhost:8081/v1/alerts 2>/dev/null || echo "[]")
ALERT_COUNT=$(echo "$ALERTS" | jq length 2>/dev/null || echo "0")

kill $PF_PID 2>/dev/null || true

echo "Alerts received: $ALERT_COUNT"

# Step 8: Verify incidents
echo -e "\n${YELLOW}Step 8: Checking results...${NC}"

if [ "$ALERT_COUNT" -ge 1 ]; then
    echo -e "${GREEN}✅ PASS: Alerts generated ($ALERT_COUNT)${NC}"
else
    echo -e "${RED}❌ FAIL: No alerts generated${NC}"
    exit 1
fi

# Check incidents
kubectl port-forward svc/podwatch-incident 8081:8081 -n $NAMESPACE &
PF_PID=$!
sleep 3

INCIDENTS=$(curl -s http://localhost:8081/v1/incidents 2>/dev/null || echo "[]")
INCIDENT_COUNT=$(echo "$INCIDENTS" | jq length 2>/dev/null || echo "0")

kill $PF_PID 2>/dev/null || true

if [ "$INCIDENT_COUNT" -ge 1 ]; then
    echo -e "${GREEN}✅ PASS: Incidents created ($INCIDENT_COUNT)${NC}"
else
    echo -e "${YELLOW}⚠️  WARNING: No incidents created${NC}"
fi

# Summary
echo -e "\n============================================"
echo -e "${GREEN}End-to-End Tests Completed Successfully!${NC}"
echo "============================================"
echo "Alerts: $ALERT_COUNT"
echo "Incidents: $INCIDENT_COUNT"
echo ""
echo "To explore the cluster:"
echo "  export KUBECONFIG=\"$(kind get kubeconfig-path --name=$CLUSTER_NAME)\""
echo "  kubectl get pods -A"
echo ""
echo "To access UI:"
echo "  kubectl port-forward svc/podwatch-ui 3000:80 -n $NAMESPACE"
echo "  open http://localhost:3000"
