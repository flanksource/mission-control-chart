#!/bin/bash
#
# Kratos Migration E2E Test
#
# Tests the upgrade path from Kratos v0.13.0 (current production via chart 0.32.0)
# to v26.2.0 with realistic data present in the database.
#
# Steps:
#   1. Run v0.13.0 migrations on a fresh Postgres DB, then start the v0.13.0 server.
#   2. Seed data via the Kratos API: identities with password credentials, sessions
#      (which also create session_devices), and recovery/verification flows
#      (which create tokens and courier messages).
#   3. Stop the old server and run v26.2.0 migrations on the populated DB.
#      This exercises schema changes, index rebuilds, and Go-based backfill
#      migrations (e.g. populating identity_id on credential_identifiers and
#      session_devices).
#   4. Verify that backfilled columns contain no NULLs — the subsequent NOT NULL
#      constraint + FK migration would fail if they did.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

OLD_VERSION="v0.13.0"
NEW_VERSION="v26.2.0"

cleanup() {
    echo "==> Cleaning up..."
    docker compose down -v --remove-orphans 2>/dev/null || true
}
trap cleanup EXIT

echo "==> Kratos migration test: ${OLD_VERSION} → ${NEW_VERSION}"
docker compose down -v --remove-orphans 2>/dev/null || true

# Step 1: Migrate + start old kratos
echo ""
echo "==> Step 1: Running ${OLD_VERSION} migration and starting server..."
docker compose up -d kratos_old_server

for i in $(seq 1 30); do
    if curl -sf http://localhost:4434/admin/health/alive > /dev/null 2>&1; then
        break
    fi
    if [ "$i" -eq 30 ]; then
        echo "✗ Kratos ${OLD_VERSION} failed to start"
        docker compose logs kratos_old_server
        exit 1
    fi
    sleep 1
done
echo "✓ Kratos ${OLD_VERSION} is running"

# Step 2: Seed data via Kratos API
echo ""
echo "==> Step 2: Seeding identities, sessions, recovery & verification flows..."

create_identity() {
    local email="$1" first="$2" last="$3" password="$4"
    curl -sf -X POST http://localhost:4434/admin/identities \
        -H "Content-Type: application/json" \
        -d "{
            \"schema_id\": \"default\",
            \"traits\": {\"email\": \"$email\", \"name\": {\"first\": \"$first\", \"last\": \"$last\"}},
            \"credentials\": {\"password\": {\"config\": {\"password\": \"$password\"}}}
        }" | jq -r '.id'
}

login_user() {
    local email="$1" password="$2"
    local flow_id
    flow_id=$(curl -sf http://localhost:4433/self-service/login/api | jq -r '.id')
    curl -sf -X POST "http://localhost:4433/self-service/login?flow=${flow_id}" \
        -H "Content-Type: application/json" \
        -d "{\"method\":\"password\",\"identifier\":\"$email\",\"password\":\"$password\"}" > /dev/null
}

trigger_recovery() {
    local email="$1"
    local flow_id
    flow_id=$(curl -sf http://localhost:4433/self-service/recovery/api | jq -r '.id')
    curl -sf -X POST "http://localhost:4433/self-service/recovery?flow=${flow_id}" \
        -H "Content-Type: application/json" \
        -d "{\"method\":\"link\",\"email\":\"$email\"}" > /dev/null
}

trigger_verification() {
    local email="$1"
    local flow_id
    flow_id=$(curl -sf http://localhost:4433/self-service/verification/api | jq -r '.id')
    curl -sf -X POST "http://localhost:4433/self-service/verification?flow=${flow_id}" \
        -H "Content-Type: application/json" \
        -d "{\"method\":\"link\",\"email\":\"$email\"}" > /dev/null
}

# Create identities
create_identity "user1@example.com" "User" "One" "Password123!"
create_identity "user2@example.com" "User" "Two" "Password456!"
create_identity "admin@local" "Admin" "User" "AdminPass123!"

# Create sessions (populates sessions + session_devices)
login_user "user1@example.com" "Password123!"
login_user "user2@example.com" "Password456!"
login_user "admin@local" "AdminPass123!"
login_user "user1@example.com" "Password123!" # second session for user1

# Trigger recovery + verification flows (populates tokens, courier messages)
trigger_recovery "user1@example.com"
trigger_verification "user2@example.com"

echo "✓ Seeded 3 identities, 4 sessions, recovery & verification flows"

# Step 3: Stop old server, run new migration
echo ""
echo "==> Step 3: Stopping ${OLD_VERSION} and running ${NEW_VERSION} migration..."
docker compose stop kratos_old_server

docker compose run --rm kratos_new_migration
echo "✓ Kratos ${NEW_VERSION} migration succeeded"

# Step 4: Verify backfilled data
echo ""
echo "==> Step 4: Verifying backfilled data..."

NULL_CREDENTIAL_IDS=$(docker compose exec -T postgres psql -U postgres -d kratos_migration_test -tAc \
    "SELECT COUNT(*) FROM identity_credential_identifiers WHERE identity_id IS NULL")
if [ "$NULL_CREDENTIAL_IDS" != "0" ]; then
    echo "✗ Found ${NULL_CREDENTIAL_IDS} credential identifiers with NULL identity_id"
    exit 1
fi

NULL_DEVICE_IDS=$(docker compose exec -T postgres psql -U postgres -d kratos_migration_test -tAc \
    "SELECT COUNT(*) FROM session_devices WHERE identity_id IS NULL")
if [ "$NULL_DEVICE_IDS" != "0" ]; then
    echo "✗ Found ${NULL_DEVICE_IDS} session devices with NULL identity_id"
    exit 1
fi

echo "✓ All backfill columns populated correctly"
echo ""
echo "==> ✓ Migration test passed: ${OLD_VERSION} → ${NEW_VERSION}"
