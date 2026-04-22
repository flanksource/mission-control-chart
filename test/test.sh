#!/bin/bash

function header () {
    echo "........................................................................................................................"
    echo "==> $1"
    echo "........................................................................................................................"
}

kubectl apply -f topologies/mission-control.yaml
kubectl get topology mission-control -o yaml

# Wait for topology to reconcile
# The Topology object Ready state is not updated when the topology is working,
# so we need to poll the database until the components are updated. A successful
# topology run in an empty database will result in components having a non-zero
# number of records. Times out after 5 minutes.
if [[ $(uname -s) == "Darwin" ]]; then
    END_TIME=$(date -v +5M +%s)
else
    TIMEOUT="5 minute"
    END_TIME=$(date -ud "$TIMEOUT" +%s)
fi
SUCCESS=0

while [[ $(date -u +%s) -le $END_TIME ]]
do
    echo "Waiting for toplogy to reconcile..."
    COUNT=$(kubectl -n default exec postgresql-0 -- bash -c 'psql $DB_URL -c "SELECT COUNT(*) FROM components;"'| grep -A 2 "count" | grep "0" -c| grep -A 2 "count" | grep -c " 0$")
    if [[ "$COUNT" != "0" ]]; then
        SUCCESS=1
        echo "Topology reconciled successfully"
        break
    fi
    sleep 5
done

header "Topology"
kubectl describe topology mission-control

header "Events of namespace default"
kubectl -n default get events --sort-by=.metadata.creationTimestamp

header "APM-hub Logs"
kubectl -n default logs -l app.kubernetes.io/name=apm-hub --tail 100

header "Canary-checker Logs"
kubectl -n default logs -l app.kubernetes.io/name=canary-checker --tail 100

header "Config-DB Logs"
kubectl -n default logs -l app.kubernetes.io/name=config-db --tail 100

header "Incident Manager UI Logs"
kubectl -n default logs -l app.kubernetes.io/name=incident-manager-ui --tail 100

header "Kratos Logs"
kubectl -n default logs -l app.kubernetes.io/name=kratos --tail 100

header "Mission Control Logs"
kubectl -n default logs -l app.kubernetes.io/name=mission-control,app.kubernetes.io/component=mission-control --tail 100

if [[ "$SUCCESS" == "1" ]]; then
    exit 1
fi
exit 0
