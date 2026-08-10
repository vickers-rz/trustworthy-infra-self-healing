# Trustworthy Infra Self-Healing

> **Reason probabilistically. Authorize deterministically.**

A research-oriented, Go-first control plane for **trustworthy, transparent, explainable infrastructure self-healing**.

The project explores a simple but demanding question:

> How can probabilistic AI safely assist deterministic production infrastructure without becoming an unrestricted production operator?

The answer is not “give an LLM shell access.” This repository separates **observation, evidence, diagnosis, proposal, trusted execution context, risk classification, authorization, execution, verification, rollback, and learning** into explicit layers with independent safety boundaries.

## Design thesis

This project inherits a core idea from HomeMind: **learning authority is not actuation authority**.

For infrastructure, that becomes:

> **Diagnostic Authority ≠ Remediation Authority**

An AI component may consume evidence, generate hypotheses, use RAG, rank candidate causes, estimate risk, and produce a structured remediation proposal. None of those capabilities imply permission to mutate production state.

The deterministic control plane remains the final authority.

## Three-way trust boundary

Authorization currently receives three deliberately separate inputs:

```text
probabilistic / model-owned       trusted runtime             trusted evidence plane

┌────────────────────────┐   ┌────────────────────────┐   ┌──────────────────────────┐
│ Proposal               │   │ ExecutionContext       │   │ EvidenceBundle           │
│                        │   │                        │   │                          │
│ hypothesis             │   │ authenticated actor    │   │ source / collector       │
│ bundle/item IDs only   │   │ granted authority      │   │ timestamps / freshness   │
│ estimated_risk         │   │ approval provenance    │   │ trust metadata           │
│ typed action           │   │ operator override      │   │ integrity metadata       │
│ preconditions          │   │ policy version         │   │ missing evidence         │
│ verification           │   │ environment            │   │ claim relations          │
│ typed rollback         │   │ decision_time          │   │                          │
└───────────┬────────────┘   └───────────┬────────────┘   └────────────┬─────────────┘
            │                            │                             │
            └────────────────────────────┼─────────────────────────────┘
                                         ▼
                              Deterministic Remediation Guard
                                         │
                              risk + freshness + policy
                                         ▼
                                   Semantic Executor
```

The proposal cannot:

- approve itself;
- grant itself authority;
- disable an operator override;
- choose the effective authorization risk;
- rewrite evidence source/collector identity;
- rewrite evidence timestamps or freshness;
- invent a trusted evidence item merely by embedding one in model output.

It can only reference an existing `evidence_bundle_id` and a bounded set of `evidence_ids`. The Guard resolves those references against the separately supplied trusted bundle.

## Evidence contract

Evidence is data with provenance, not free-form context pasted into a prompt.

Each evidence item may carry:

- stable ID and kind;
- URI, source system and collector identity;
- subject/resource;
- untrusted summary content;
- observed and collected timestamps;
- optional observation window;
- optional freshness deadline;
- source trust metadata;
- optional SHA-256 integrity digest.

Important distinctions:

- `summary` is **untrusted evidence content**, never an instruction;
- `trust` describes provenance quality, not hypothesis probability;
- evidence weighting/ranking belongs to diagnosis logic, not the evidence object;
- `fresh`, `stale`, and `freshness unknown` are distinct states;
- supporting and contradicting relations are explicit rather than hidden inside generated prose;
- missing telemetry can be represented explicitly instead of silently disappearing.

### Deterministic freshness

Freshness is evaluated against trusted `ExecutionContext.decision_time`, not an implicit wall clock. This makes the decision replayable during audit.

For mutable remediation, the initial policy requires at least one selected **fresh operational evidence** item. Operational evidence such as metrics, logs, traces, Kubernetes state/events, changes, topology, and operator annotations must have known freshness and still be fresh at decision time.

Runbooks and historical incidents may accompany a decision as reference context, but cannot by themselves establish current operational state.

## Core principles

1. **Observe freely.** Read-only telemetry and evidence collection are broadly available.
2. **Preserve provenance.** Evidence retains source, collector, time, subject, freshness semantics, and integrity metadata when available.
3. **Separate evidence authority from proposal authority.** Models reference evidence; they do not author trusted provenance for authorization.
4. **Reason probabilistically.** ML/LLM components may express uncertainty, hypotheses, confidence, alternatives, and advisory risk estimates.
5. **Propose structurally.** AI outputs typed remediation proposals, never arbitrary shell commands.
6. **Classify risk deterministically.** Effective operational risk is derived from trusted semantic action rules, not accepted from model output.
7. **Authorize from trusted context.** Identity, authority, approval, override, environment, policy version, and decision time are supplied outside the proposal.
8. **Execute minimally.** Prefer the smallest reversible intervention that can validate the hypothesis.
9. **Verify independently.** The component proposing a repair does not get to certify success.
10. **Fail closed on ambiguity.** Unknown actions, malformed bundles, stale operational evidence, missing rollback, or policy ambiguity reduce authority rather than expand it.

## Non-goals

This project is **not** intended to be:

- an unrestricted autonomous `kubectl`/SSH agent;
- an LLM wrapper around shell execution;
- a system where model confidence overrides policy;
- a system that trusts an AI-provided `human_approved: true` flag;
- a system that accepts model-declared risk as authorization truth;
- a system that lets a model declare its own evidence `trust: high` or freshness and then treats that as trusted input;
- a system that treats retrieved text as instructions;
- a mechanism for silently expanding its own production permissions;
- an opaque “AI fixed it” black box.

## High-level architecture

```text
Metrics / Logs / Traces / K8s Events / Git / Runbooks / Incidents
                              │
                              ▼
                       Evidence Plane
              normalize / provenance / freshness
                              │
                    trusted EvidenceBundle
                              │
              ┌───────────────┴───────────────┐
              │                               │
              ▼                               ▼
       Diagnosis / RAG / ML             Remediation Guard
              │                               ▲
        typed Proposal                         │
              │                      trusted ExecutionContext
              └───────────────────────┬───────┘
                                      ▼
                         Deterministic Risk Classifier
                                      │
                              effective risk
                                      ▼
                             Freshness / Policy
                                      │
                              allow / deny / escalate
                                      ▼
                              Semantic Executor
                                      │
                              dry-run / canary
                                      ▼
                               Production Action
                                      │
                       ┌──────────────┴──────────────┐
                       ▼                             ▼
                 Independent                   Verification
                  Watchdog                          │
                       │                             │
                       └────── abort/rollback ──────┘
                                      │
                                      ▼
                                  Audit Log
                                      │
                                      ▼
                               Outcome / Learning
```

## Authority model

### Execution authority

| Level | Capability |
|---|---|
| `L0 OBSERVE` | read telemetry and state |
| `L1 DIAGNOSE` | produce root-cause hypotheses |
| `L2 PROPOSE` | generate structured remediation proposals |
| `L3 PREPARE` | render a bounded execution plan without applying it |
| `L4 DELEGATED` | execute pre-authorized low-risk playbooks |
| `L5 APPROVAL` | execute only with explicit human approval provenance |
| `DENY` | action is never executable |

### Operational risk / blast radius

| Class | Meaning | Example |
|---|---|---|
| `R0` | observation only | query metrics/logs |
| `R1` | ephemeral/single workload | restart one stateless workload |
| `R2` | service-local reversible change | rollback one deployment |
| `R3` | multi-service/stateful impact | database failover |
| `R4` | business/security critical | IAM/network/schema mutation |
| `R5` | catastrophic or irreversible | destructive storage/database action |

Authority and effective risk are evaluated together. A high authority level does **not** imply permission for a high-risk action.

## Proposal references evidence; it does not own evidence

A trusted evidence plane may produce:

```yaml
evidence_bundle:
  id: eb_payment_001
  incident_id: inc_payment_5xx
  created_at: 2026-08-10T04:12:03Z
  items:
    - id: ev_error_rate
      kind: metric
      uri: prometheus://payment-api/error-rate?window=5m
      source: prometheus
      collector: infraheal-prometheus-adapter/v0
      subject: k8s://payments/deployment/payment-api
      summary: 5xx rose after revision 193 rollout
      observed_at: 2026-08-10T04:12:00Z
      collected_at: 2026-08-10T04:12:02Z
      fresh_until: 2026-08-10T04:14:00Z
      trust: high
```

A model/planner may then produce only references plus typed intent:

```yaml
proposal:
  id: rem_20260810_001
  incident_id: inc_payment_5xx
  evidence_bundle_id: eb_payment_001
  evidence_ids:
    - ev_error_rate
  estimated_risk: R1  # advisory, intentionally under-reported here
  hypothesis:
    id: hyp_bad_deployment
    type: unhealthy_new_deployment
    confidence: 0.87
  action:
    type: rollback_deployment
    rollback_deployment:
      namespace: payments
      deployment: payment-api
      from_revision: 193
      to_revision: 192
  rollback:
    type: rollback_deployment
    rollback_deployment:
      namespace: payments
      deployment: payment-api
      from_revision: 192
      to_revision: 193
  preconditions:
    - previous revision is available
    - migration is backward-compatible
  verification:
    - error_rate_5m < 1%
```

Trusted runtime state is supplied separately:

```yaml
execution_context:
  authority: L5_APPROVAL
  human_approval:
    approval_id: CHG-1234
    approved_by: operator@example.com
  operator_override: false
  policy_version: mvp-v3
  environment: staging
  actor_id: controlplane:staging
  decision_time: 2026-08-10T04:13:00Z
```

The deterministic classifier derives `rollback_deployment` as **R2**. The under-reported `R1` estimate therefore fails closed. Independently, the Guard verifies that `ev_error_rate` exists in `eb_payment_001` and is fresh at the trusted decision time.

## Typed semantic action catalogue

The MVP uses a closed action union rather than `map[string]string` command parameters.

Current types include:

- `restart_workload`;
- `rollback_deployment`;
- `scale_workload`;
- `drain_node`;
- `uncordon_node`;
- `failover_database`.

Each action has its own typed payload and validation rules. Rollback/compensation is itself a typed semantic action.

## Kubernetes substrate

The first real Kubernetes control loop is deliberately observe-only:

- `infraheal.io/v1alpha1` `HealingPolicy` CRD;
- Deployment observation and status conditions;
- controller-runtime cache field index for event mapping;
- generated CRD/deepcopy/RBAC with zero-diff CI verification;
- target-side Deployment RBAC is only `get/list/watch`;
- Kind e2e runs the controller under its real restricted ServiceAccount and proves target patch is forbidden.

No LLM, RAG, or remediation executor is connected to this controller yet.

## Repository map

```text
cmd/controlplane/        deterministic decision demo
cmd/controller/          observe-only Kubernetes controller
api/v1alpha1/            HealingPolicy Kubernetes API
internal/domain/         proposal/action/context/evidence domain models
internal/evidence/       deterministic evidence/freshness policy
internal/risk/           deterministic effective-risk classifier
internal/policy/         remediation guard
internal/executor/       semantic executor boundary
internal/controller/     observe-only reconciliation logic
docs/EVIDENCE_MODEL.md   evidence-plane trust semantics
docs/MANIFESTO.md        design principles
docs/ARCHITECTURE.md     architecture and control flow
docs/THREAT_MODEL.md     failure and abuse model
docs/REMEDIATION_POLICY.md policy semantics
docs/LEARNING_ROADMAP.md research + learning plan
config/                  CRD, generated RBAC, manager and examples
test/e2e/                restricted-RBAC Kind safety test
examples/                separate evidence/proposal/context example
```

## Current milestones

Completed foundations include:

- closed typed remediation action schema;
- trusted execution context separated from planner output;
- deterministic effective-risk classification;
- provenance-bearing `EvidenceBundle` model;
- Proposal → trusted bundle ID/reference binding;
- trusted/replayable decision time;
- deterministic operational-evidence freshness policy;
- fail-closed behavior for stale/unknown operational evidence;
- mandatory typed rollback where required;
- observe-only `HealingPolicy` Kubernetes controller;
- controller-gen-owned CRD/deepcopy/RBAC artifacts with zero-diff CI verification;
- restricted ServiceAccount Kind test proving the controller cannot patch Deployments;
- indexed Deployment-to-`HealingPolicy` event routing;
- mock semantic executor;
- unit tests for trust-boundary bypass attempts.

## Development

```bash
go test ./...
go run ./cmd/controlplane
```

Controller-specific targets:

```bash
make verify-generated
make install-crd
make run-controller
make sample
```

## Project status

Early research and engineering scaffold. The repository is deliberately safety-first: LLM/RAG diagnosis will be connected only after evidence adapters, source trust, audit, simulation, watchdog, rollback, and verification mechanisms are independently testable.

## License

Apache-2.0.
