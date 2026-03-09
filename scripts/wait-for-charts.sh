#!/bin/bash
set -e

# Script to wait for Helm chart dependencies to become available in repositories
# This ensures that newly published charts are available before attempting to package dependent charts

CHART_FILES=("$@")
if [ ${#CHART_FILES[@]} -eq 0 ]; then
  echo "Usage: $0 <Chart.yaml> [<Chart.yaml> ...]"
  exit 1
fi

TIMEOUT=300  # 5 minutes timeout
POLL_INTERVAL=10  # Check every 10 seconds

echo "=========================================="
echo "Checking chart dependency availability"
echo "=========================================="

# Function to check if a chart version exists in a repository
check_chart_exists() {
  local chart_name=$1
  local chart_version=$2
  local repo_url=$3

  # For flanksource charts, check the index
  if [[ "$repo_url" == *"flanksource.github.io/charts"* ]]; then
    # Fetch the chart index
    local index_url="${repo_url}/index.yaml"
    local chart_info
    # Use || true to prevent curl failures from exiting the script
    chart_info=$(curl -sSf "$index_url" 2>/dev/null | yq eval ".entries.\"${chart_name}\"[] | select(.version == \"${chart_version}\")" - 2>/dev/null || true)

    if [ -n "$chart_info" ]; then
      return 0
    else
      return 1
    fi
  else
    # For other repositories, try helm search
    helm repo update > /dev/null 2>&1 || true
    if helm search repo --version "$chart_version" --regexp ".*/${chart_name}$" 2>/dev/null | grep -q "$chart_version"; then
      return 0
    else
      return 1
    fi
  fi
}

# Extract dependencies from Chart.yaml files
missing_charts=()

for chart_file in "${CHART_FILES[@]}"; do
  if [ ! -f "$chart_file" ]; then
    echo "Error: Chart file not found: $chart_file"
    exit 1
  fi

  echo ""
  echo "Processing: $chart_file"

  # Extract dependencies using yq
  deps_count=$(yq eval '.dependencies | length' "$chart_file")

  if [ "$deps_count" -eq 0 ] || [ "$deps_count" == "null" ]; then
    echo "  No dependencies found"
    continue
  fi

  for ((i=0; i<deps_count; i++)); do
    name=$(yq eval ".dependencies[$i].name" "$chart_file")
    version=$(yq eval ".dependencies[$i].version" "$chart_file")
    repository=$(yq eval ".dependencies[$i].repository" "$chart_file")

    if [ "$name" == "null" ] || [ "$version" == "null" ]; then
      continue
    fi

    # Skip version ranges (e.g., ">= 0.0.20")
    if [[ "$version" =~ ^[\>\<\=\~\^] ]] || [[ "$version" == "*" ]]; then
      echo "  Skipping version range: $name@$version from $repository"
      continue
    fi

    echo "  Checking: $name@$version from $repository"

    # Add to list for verification
    missing_charts+=("$name|$version|$repository")
  done
done

# Remove duplicates
missing_charts=($(printf '%s\n' "${missing_charts[@]}" | sort -u))

if [ ${#missing_charts[@]} -eq 0 ]; then
  echo ""
  echo "No chart dependencies to verify"
  exit 0
fi

echo ""
echo "=========================================="
echo "Waiting for ${#missing_charts[@]} chart dependencies to be available..."
echo "Timeout: ${TIMEOUT}s, Poll interval: ${POLL_INTERVAL}s"
echo "=========================================="

start_time=$(date +%s)
all_available=false

while true; do
  current_time=$(date +%s)
  elapsed=$((current_time - start_time))

  if [ $elapsed -ge $TIMEOUT ]; then
    echo ""
    echo "❌ Timeout reached after ${TIMEOUT}s"
    echo "The following charts are still not available:"
    for chart_info in "${missing_charts[@]}"; do
      IFS='|' read -r name version repository <<< "$chart_info"
      echo "  - $name@$version"
    done
    exit 1
  fi

  remaining_charts=()
  available_count=0

  for chart_info in "${missing_charts[@]}"; do
    IFS='|' read -r name version repository <<< "$chart_info"

    if check_chart_exists "$name" "$version" "$repository"; then
      echo "✓ Available: $name@$version"
      ((available_count++))
    else
      remaining_charts+=("$chart_info")
    fi
  done

  if [ ${#remaining_charts[@]} -eq 0 ]; then
    all_available=true
    break
  fi

  missing_charts=("${remaining_charts[@]}")

  echo ""
  echo "Progress: $available_count available, ${#missing_charts[@]} remaining"
  echo "Waiting ${POLL_INTERVAL}s before next check... (elapsed: ${elapsed}s)"
  sleep $POLL_INTERVAL
done

if [ "$all_available" = true ]; then
  echo ""
  echo "=========================================="
  echo "✅ All chart dependencies are available!"
  echo "=========================================="
  exit 0
fi
