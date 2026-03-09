#!/bin/bash
set -e

# Wait for Helm chart dependencies to become available, then run helm dependency update.
# Optimized: fetches each repo index once per poll cycle (not per chart).

CHART_DIRS=("$@")
if [ ${#CHART_DIRS[@]} -eq 0 ]; then
  echo "Usage: $0 <chart-dir> [<chart-dir> ...]"
  echo "Example: $0 ./chart ./agent-chart"
  exit 1
fi

TIMEOUT=${TIMEOUT:-300}
POLL_INTERVAL=${POLL_INTERVAL:-10}
INDEX_CACHE_DIR=$(mktemp -d)
trap 'rm -rf "$INDEX_CACHE_DIR"' EXIT

for dir in "${CHART_DIRS[@]}"; do
  if [ ! -f "$dir/Chart.yaml" ]; then
    echo "Error: $dir/Chart.yaml not found"
    exit 1
  fi
done

echo "=========================================="
echo "Checking chart dependency availability"
echo "=========================================="

# Extract all pinned dependencies across all Chart.yaml files, deduplicated
missing_charts=()
declare -A seen

for dir in "${CHART_DIRS[@]}"; do
  chart_file="$dir/Chart.yaml"
  echo ""
  echo "Processing: $chart_file"

  deps=$(yq eval -o=json '.dependencies // []' "$chart_file")
  count=$(echo "$deps" | yq eval 'length' -)

  for ((i=0; i<count; i++)); do
    name=$(echo "$deps" | yq eval ".[$i].name" -)
    version=$(echo "$deps" | yq eval ".[$i].version" -)
    repository=$(echo "$deps" | yq eval ".[$i].repository" -)

    [ "$name" == "null" ] || [ "$version" == "null" ] && continue

    # Skip non-flanksource repos
    if [[ "$repository" != *"flanksource.github.io/charts"* ]]; then
      echo "  Skipping non-flanksource: $name@$version ($repository)"
      continue
    fi

    # Skip version ranges
    if [[ "$version" =~ ^[\>\<\=\~\^] ]] || [[ "$version" == "*" ]]; then
      echo "  Skipping range: $name@$version"
      continue
    fi

    key="$name|$version|$repository"
    if [ -z "${seen[$key]}" ]; then
      seen[$key]=1
      missing_charts+=("$key")
      echo "  Checking: $name@$version"
    fi
  done
done

if [ ${#missing_charts[@]} -eq 0 ]; then
  echo ""
  echo "No chart dependencies to verify"
  for dir in "${CHART_DIRS[@]}"; do
    echo "Running: helm dependency update $dir"
    helm dependency update "$dir"
  done
  exit 0
fi

# Fetch a repo index once, cache it for this poll cycle
fetch_index() {
  local repo_url=$1
  local cache_key
  cache_key=$(echo "$repo_url" | md5sum | cut -d' ' -f1 2>/dev/null || echo "$repo_url" | md5 -q 2>/dev/null || echo "$repo_url" | tr -dc 'a-zA-Z0-9')
  local cache_file="$INDEX_CACHE_DIR/$cache_key"

  if [ ! -f "$cache_file" ]; then
    curl -sSf "${repo_url}/index.yaml" > "$cache_file" 2>/dev/null || { rm -f "$cache_file"; return 1; }
  fi
  echo "$cache_file"
}

clear_index_cache() {
  rm -f "$INDEX_CACHE_DIR"/*
}

check_chart_exists() {
  local name=$1 version=$2 repo_url=$3
  local cache_file
  cache_file=$(fetch_index "$repo_url") || return 1
  yq eval -e ".entries.\"${name}\"[] | select(.version == \"${version}\")" "$cache_file" > /dev/null 2>&1
}

get_latest_version_info() {
  local name=$1 repo_url=$2
  local cache_file
  cache_file=$(fetch_index "$repo_url" 2>/dev/null) || { echo "unknown"; return; }
  yq eval ".entries.\"${name}\"[0] | .version + \" (\" + .created + \")\"" "$cache_file" 2>/dev/null || echo "unknown"
}

echo ""
echo "=========================================="
echo "Waiting for ${#missing_charts[@]} chart dependencies..."
echo "Timeout: ${TIMEOUT}s, Poll interval: ${POLL_INTERVAL}s"
echo "=========================================="

start_time=$(date +%s)

while true; do
  elapsed=$(( $(date +%s) - start_time ))

  if [ $elapsed -ge $TIMEOUT ]; then
    echo ""
    echo "Timeout reached after ${TIMEOUT}s. Still missing:"
    for entry in "${missing_charts[@]}"; do
      IFS='|' read -r name version repository <<< "$entry"
      latest=$(get_latest_version_info "$name" "$repository")
      echo "  - $name@$version (latest: $latest)"
    done
    exit 1
  fi

  clear_index_cache
  remaining=()

  for entry in "${missing_charts[@]}"; do
    IFS='|' read -r name version repository <<< "$entry"
    if check_chart_exists "$name" "$version" "$repository"; then
      echo "  Available: $name@$version"
    else
      remaining+=("$entry")
    fi
  done

  if [ ${#remaining[@]} -eq 0 ]; then
    break
  fi

  missing_charts=("${remaining[@]}")
  echo ""
  echo "  ${#missing_charts[@]} remaining (${elapsed}s elapsed):"
  for entry in "${missing_charts[@]}"; do
    IFS='|' read -r name version repository <<< "$entry"
    latest=$(get_latest_version_info "$name" "$repository")
    echo "    - waiting for $name@$version (latest: $latest)"
  done
  sleep $POLL_INTERVAL
done

echo ""
echo "All chart dependencies are available!"
echo "=========================================="

for dir in "${CHART_DIRS[@]}"; do
  echo "Running: helm dependency update $dir"
  helm dependency update "$dir"
done
