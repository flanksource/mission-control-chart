#!/bin/bash
set -euo pipefail

# Mirror dependent images to public.ecr.aws/k4y9r6y5/
# This script is called as part of the chart release process.
# Usage: scripts/mirror-images.sh [chart-dir]

CHART_DIR="${1:-chart}"
REGISTRY="public.ecr.aws/k4y9r6y5"

mirror_image() {
  local src="$1"
  local dst="${REGISTRY}/$2"
  echo "Mirroring ${src} -> ${dst}"
  crane copy "${src}" "${dst}"
}

# busybox — init container in PostgreSQL StatefulSet
mirror_image "busybox:latest" "busybox:latest"

# PostgreSQL — flanksource-maintained PostgreSQL image
POSTGRES_TAG=$(grep -oP 'flanksource/postgres:\K[^\s"]+' "${CHART_DIR}/templates/postgres.yaml" | head -1)
mirror_image "ghcr.io/flanksource/postgres:${POSTGRES_TAG}" "postgres:${POSTGRES_TAG}"

# pganalyze collector — optional database monitoring sidecar
mirror_image "quay.io/pganalyze/collector:stable" "pganalyze-collector:stable"

# pgBouncer — connection pooler (version read from chart values)
PGBOUNCER_IMAGE=$(yq '.global.db.connectionPooler.image' "${CHART_DIR}/values.yaml")
PGBOUNCER_TAG="${PGBOUNCER_IMAGE#*:}"
mirror_image "docker.io/${PGBOUNCER_IMAGE}" "pgbouncer:${PGBOUNCER_TAG}"

# PostgREST — REST API for PostgreSQL (version read from chart values)
POSTGREST_TAG=$(yq '.externalPostgrest.tag' "${CHART_DIR}/values.yaml")
mirror_image "docker.io/postgrest/postgrest:${POSTGREST_TAG}" "postgrest:${POSTGREST_TAG}"

# kubectl — used by the optional resource cleanup job
mirror_image "docker.io/bitnami/kubectl:latest" "kubectl:latest"

echo "All images mirrored successfully."
