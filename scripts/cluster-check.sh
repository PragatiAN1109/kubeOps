
#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-default}"
DEPLOYMENT="${DEPLOYMENT:-kubeops-api}"
EXPECTED_CONTEXT="${EXPECTED_CONTEXT:-kind-kubeops}"

failures=0

log() {
    echo "[INFO] $1"
}

pass() {
    echo "[PASS] $1"
}

fail() {
    echo "[FAIL] $1"
    failures=$((failures + 1))
}

# Check required dependencies and Kubernetes connectivity
check_connectivity() {
    log "Checking Kubernetes connectivity..."

    if ! command -v kubectl >/dev/null 2>&1; then
        echo "[ERROR] kubectl is not installed"
        exit 2
    fi

    local context
    context=$(kubectl config current-context 2>/dev/null) || {
        echo "[ERROR] Unable to determine Kubernetes context"
        exit 2
    }

    if [[ "$context" != "$EXPECTED_CONTEXT" ]]; then
        echo "[ERROR] Unexpected context: $context"
        exit 2
    fi

    if ! kubectl --request-timeout=10s get --raw=/readyz >/dev/null 2>&1; then
        echo "[ERROR] Kubernetes API is unreachable or not ready"
        exit 2
    fi

    pass "Connected to $context"
}

# Check that all cluster Nodes are Ready
check_nodes() {
    log "Checking Nodes..."

    local nodes
    nodes=$(kubectl get nodes \
        -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{range .status.conditions[?(@.type=="Ready")]}{.status}{end}{"\n"}{end}') || {
        fail "Unable to retrieve Nodes"
        return
    }

    if [[ -z "$nodes" ]]; then
        fail "No Nodes found"
        return
    fi

    while IFS='|' read -r name ready; do
        [[ -z "$name" ]] && continue

        if [[ "$ready" == "True" ]]; then
            pass "Node $name is Ready"
        else
            fail "Node $name is NotReady"
        fi
    done <<< "$nodes"
}

# Check Pod phases and container readiness
check_pods() {
    log "Checking Pods in namespace $NAMESPACE..."

    local pods
    pods=$(kubectl get pods -n "$NAMESPACE" \
        -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.status.phase}{"|"}{range .status.containerStatuses[*]}{.ready}{","}{.state.waiting.reason}{","}{.state.terminated.reason}{";"}{end}{"\n"}{end}') || {
        fail "Unable to retrieve Pods"
        return
    }

    if [[ -z "$pods" ]]; then
        fail "No Pods found in namespace $NAMESPACE"
        return
    fi

    while IFS='|' read -r name phase containers; do
        [[ -z "$name" ]] && continue

        if [[ "$phase" == "Succeeded" ]]; then
            pass "Pod $name completed successfully"
        elif [[ "$phase" != "Running" ]]; then
            fail "Pod $name is $phase"
        elif [[ -z "$containers" || "$containers" == *"false,"* ]]; then
            fail "Pod $name has containers that are not Ready: $containers"
        else
            pass "Pod $name is Running and Ready"
        fi
    done <<< "$pods"
}

# Check Deployment replica availability
check_deployment() {
    log "Checking Deployment $DEPLOYMENT..."

    local desired ready available observed generation

    if ! kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" >/dev/null 2>&1; then
        fail "Deployment $DEPLOYMENT not found"
        return
    fi

    local values
    values=$(kubectl get deployment "$DEPLOYMENT" -n "$NAMESPACE" \
        -o jsonpath='{.spec.replicas}{"|"}{.status.readyReplicas}{"|"}{.status.availableReplicas}{"|"}{.status.observedGeneration}{"|"}{.metadata.generation}') || {
        fail "Unable to inspect Deployment $DEPLOYMENT"
        return
    }

    IFS='|' read -r desired ready available observed generation <<< "$values"

    desired="${desired:-1}"
    ready="${ready:-0}"
    available="${available:-0}"
    observed="${observed:-0}"

    if [[ "$observed" -lt "$generation" ]]; then
        fail "Deployment controller has not observed the latest generation"
    elif [[ "$ready" -eq "$desired" && "$available" -eq "$desired" ]]; then
        pass "Deployment $DEPLOYMENT is healthy ($ready/$desired Ready)"
    else
        fail "Deployment $DEPLOYMENT unhealthy ($ready/$desired Ready, $available Available)"
    fi
}

main() {
    echo "========================================"
    echo "       KubeOps Cluster Health Check"
    echo "========================================"

    check_connectivity
    check_nodes
    check_pods
    check_deployment

    echo "========================================"

    if [[ "$failures" -eq 0 ]]; then
        echo "[SUCCESS] All health checks passed"
        exit 0
    else
        echo "[FAILED] $failures health check(s) failed"
        exit 1
    fi
}

main "$@"
