#!/usr/bin/env bash
# Production deploy: stop, remove old images, rebuild, start.
# Run from doctorandme-backend/ on the server after git pull.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${PROJECT_DIR}"

COMPOSE=(docker compose)
NO_CACHE=false

usage() {
  cat <<'EOF'
Usage: ./scripts/deploy-production.sh [options]

Options:
  --no-cache   Rebuild images without Docker layer cache (slower, fully fresh)
  -h, --help   Show this help

Workflow:
  1. docker compose down
  2. Remove old project-built images (hyphen and legacy underscore names)
  3. docker compose build
  4. docker compose up -d
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-cache)
      NO_CACHE=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ ! -f .env ]]; then
  echo "Error: .env not found in ${PROJECT_DIR}" >&2
  exit 1
fi

PROJECT_NAME="${COMPOSE_PROJECT_NAME:-doctorandme-backend}"
SERVICES=(
  auth-service
  organization-service
  appointment-service
  auth-migrations
  organization-migrations
  appointment-migrations
)

echo "Deploying ${PROJECT_NAME} from ${PROJECT_DIR}"
echo "=================================================="

echo "Step 1/4: Stopping containers and removing compose-built images..."
"${COMPOSE[@]}" down --rmi local --remove-orphans

echo "Step 2/4: Removing legacy project images (if any)..."
  docker rmi -f "${PROJECT_NAME}-${service}:latest" 2>/dev/null || true
  docker rmi -f "${PROJECT_NAME}_${service}:latest" 2>/dev/null || true
done

docker image prune -f >/dev/null

echo "Step 3/4: Building images..."
BUILD_ARGS=(build)
if [[ "${NO_CACHE}" == "true" ]]; then
  BUILD_ARGS+=(--no-cache)
fi
"${COMPOSE[@]}" "${BUILD_ARGS[@]}"

echo "Step 4/4: Starting containers..."
"${COMPOSE[@]}" up -d --force-recreate

echo ""
echo "Deployment complete."
echo ""
"${COMPOSE[@]}" ps
echo ""
echo "Useful commands:"
echo "  docker compose logs -f"
echo "  docker compose logs -f auth-service"
