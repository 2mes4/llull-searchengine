#!/usr/bin/env bash
set -euo pipefail

IMAGE="${IMAGE:-llull-searchengine}"
TAG="${TAG:-latest}"

echo "Building $IMAGE:$TAG..."
docker build -t "$IMAGE:$TAG" -f deploy/docker/Dockerfile.server .

echo "Tagging and pushing (uncomment when ready)..."
# docker tag "$IMAGE:$TAG" "docker.io/$IMAGE:$TAG"
# docker push "docker.io/$IMAGE:$TAG"

echo "Done. Image: $IMAGE:$TAG"
