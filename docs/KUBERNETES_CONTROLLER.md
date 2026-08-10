# Kubernetes Controller — Observe-Only Milestone

This is the first real Kubernetes control loop in the project. It is intentionally **not a self-healing executor yet**.

The purpose of this milestone is to establish the substrate that later evidence, diagnosis and remediation systems will depend on, while preserving a hard safety boundary: the controller can observe target workloads but cannot mutate them.

## API

The experimental API group is:

```text
infraheal.io/v1alpha1
```

The first custom resource is `HealingPolicy`.

```yaml
apiVersion: infraheal.io/v1alpha1
kind: HealingPolicy
metadata:
  name: demo-api
  namespace: default
spec:
  mode: Observe
  observeIntervalSeconds: 30
  target:
    name: demo-api
```

The MVP target type is an `apps/v1 Deployment`.

## Reconciliation

```text
HealingPolicy event ───────────────┐
                                   │
Deployment event ── target lookup ─┼──► Reconcile
                                   │
periodic requeue ──────────────────┘
                                      │
                                      ▼
                              GET Deployment
                                /          \
                            found          missing
                              │               │
                              ▼               ▼
                       observe replicas   mark unresolved
                              │               │
                              └──────┬────────┘
                                     ▼
                           UPDATE HealingPolicy/status
```

The controller records:

- whether the target was resolved;
- desired replicas;
- available replicas;
- ready replicas;
- updated replicas;
- observation timestamp;
- Kubernetes Conditions describing target resolution, observed health and the observe-only safety boundary.

## Condition semantics

`LastTransitionTime` represents a **condition status transition**, not a reconciliation timestamp. The controller uses Kubernetes `meta.SetStatusCondition`, so repeated observations that leave a condition in the same `True`, `False` or `Unknown` state preserve its transition time.

`LastObservedTime` is separate and may advance on every successful observation. This keeps state-transition history distinct from the observation heartbeat.

## Hard safety boundary

The target-side ClusterRole grants only:

```text
get
list
watch
```

on Deployments.

There is deliberately no target-side:

```text
create
update
patch
delete
deletecollection
```

The controller may update only the `HealingPolicy/status` subresource.

This matters because the safety property is enforced at more than one layer:

1. the controller code contains no remediation path;
2. the CRD exposes only `mode: Observe` in this milestone;
3. RBAC denies target mutation even if a future code defect attempted one.

## Local development

Install the CRD:

```bash
make install-crd
```

Run the controller against the current kubeconfig:

```bash
make run-controller
```

Apply a sample policy:

```bash
make sample
```

Inspect status:

```bash
kubectl get healingpolicy demo-api -o yaml
```

## Tests

The fake-client test suite verifies:

- healthy Deployment observation;
- missing-target behavior;
- Deployment event to HealingPolicy mapping;
- target Deployment spec remains unchanged;
- stable condition state preserves `LastTransitionTime` across repeated reconciliation;
- `LastObservedTime` advances independently.

## What this milestone does not prove yet

A fake client is not an API server. Before Phase 1.5 is considered complete, the project still needs a reproducible Kind integration test that proves:

- the CRD is accepted by a real Kubernetes API server;
- the manager watches and reconciles real objects;
- status subresource writes behave correctly;
- Deployment events trigger observation;
- the controller service account cannot patch or update Deployments;
- restart/reconciliation behavior remains safe.

That Kind test is the next milestone.
