#!/usr/bin/env bash
set -euo pipefail

controller_identity="system:serviceaccount:infraheal-system:infraheal-controller"

wait_jsonpath() {
  local resource="$1"
  local jsonpath="$2"
  local expected="$3"
  local timeout_seconds="${4:-90}"
  local start
  start="$(date +%s)"

  while true; do
    local value
    value="$(kubectl get ${resource} -o "jsonpath=${jsonpath}" 2>/dev/null || true)"
    if [[ "${value}" == "${expected}" ]]; then
      return 0
    fi
    if (( $(date +%s) - start >= timeout_seconds )); then
      echo "timed out waiting for ${resource} ${jsonpath}=${expected}; last value=${value}" >&2
      kubectl get healingpolicies.infraheal.io -A -o yaml || true
      kubectl get deployments -A || true
      kubectl logs -n infraheal-system deployment/infraheal-controller || true
      return 1
    fi
    sleep 2
  done
}

echo "==> installing CRD and restricted controller RBAC"
kubectl apply -f config/crd/bases/infraheal.io_healingpolicies.yaml
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/manager/controller.yaml
kubectl rollout status -n infraheal-system deployment/infraheal-controller --timeout=90s

echo "==> proving target-side RBAC is read-only"
# kubectl auth can-i intentionally returns a non-zero process status when the
# authorization answer is "no". Capture the textual answer so set -e does not
# turn an expected denial into a test harness failure.
can_get="$(kubectl auth can-i get deployments.apps --as="${controller_identity}" --all-namespaces || true)"
can_patch="$(kubectl auth can-i patch deployments.apps --as="${controller_identity}" --all-namespaces || true)"
echo "controller get deployments: ${can_get}"
echo "controller patch deployments: ${can_patch}"
if [[ "${can_get}" != "yes" ]]; then
  echo "controller identity unexpectedly cannot read Deployments" >&2
  exit 1
fi
if [[ "${can_patch}" != "no" ]]; then
  echo "controller identity unexpectedly can patch Deployments" >&2
  exit 1
fi

kubectl apply -f test/e2e/fixtures/demo-deployment.yaml
kubectl apply -f config/samples/infraheal_v1alpha1_healingpolicy.yaml

wait_jsonpath "healingpolicy/demo-api" "{.status.targetFound}" "true"
wait_jsonpath "healingpolicy/demo-api" "{.status.desiredReplicas}" "1"

initial_evidence_id="$(kubectl get healingpolicy/demo-api -o jsonpath='{.status.lastEvidenceID}')"
initial_digest="$(kubectl get healingpolicy/demo-api -o jsonpath='{.status.lastEvidenceDigestSHA256}')"
if [[ ! "${initial_digest}" =~ ^[0-9a-f]{64}$ ]]; then
  echo "expected 64-character lowercase SHA-256 digest, got ${initial_digest}" >&2
  exit 1
fi
if [[ "${initial_evidence_id}" != "ev_k8s_${initial_digest}" ]]; then
  echo "evidence ID is not bound to its digest: id=${initial_evidence_id} digest=${initial_digest}" >&2
  exit 1
fi

echo "==> proving Deployment events drive a new evidence-backed observation"
kubectl scale deployment/demo-api --replicas=2
wait_jsonpath "healingpolicy/demo-api" "{.status.desiredReplicas}" "2"

updated_evidence_id="$(kubectl get healingpolicy/demo-api -o jsonpath='{.status.lastEvidenceID}')"
updated_digest="$(kubectl get healingpolicy/demo-api -o jsonpath='{.status.lastEvidenceDigestSHA256}')"
if [[ ! "${updated_digest}" =~ ^[0-9a-f]{64}$ ]]; then
  echo "updated evidence digest is invalid: ${updated_digest}" >&2
  exit 1
fi
if [[ "${updated_evidence_id}" != "ev_k8s_${updated_digest}" ]]; then
  echo "updated evidence ID is not bound to its digest" >&2
  exit 1
fi
if [[ "${updated_evidence_id}" == "${initial_evidence_id}" ]]; then
  echo "Deployment state changed but evidence identity did not" >&2
  exit 1
fi

if kubectl --as="${controller_identity}" patch deployment/demo-api --type merge -p '{"spec":{"replicas":3}}'; then
  echo "restricted controller identity was able to mutate the target Deployment" >&2
  exit 1
fi

actual_replicas="$(kubectl get deployment/demo-api -o jsonpath='{.spec.replicas}')"
if [[ "${actual_replicas}" != "2" ]]; then
  echo "target Deployment changed unexpectedly; replicas=${actual_replicas}" >&2
  exit 1
fi

echo "==> observe-only Kind evidence test passed"
kubectl get healingpolicy/demo-api -o yaml
