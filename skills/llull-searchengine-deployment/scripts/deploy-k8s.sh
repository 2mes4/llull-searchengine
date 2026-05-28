#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-llull}"

echo "Deploying Llull to Kubernetes namespace $NAMESPACE..."

kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

echo "Waiting for deployment to be ready..."
kubectl -n "$NAMESPACE" rollout status deployment/llull --timeout=120s

echo "Forwarding port 8080..."
kubectl -n "$NAMESPACE" port-forward svc/llull 8080:8080 &
echo "API available at http://localhost:8080"
