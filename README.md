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

An AI component may read logs, metrics, traces, deployment history, runbooks, incident reports, and topology; generate hypotheses; retrieve evidence with RAG; rank candidate causes; estimate risk; and produce a structured remediation proposal. None of those capabilities imply permission to mutate production state.

The deterministic control plane remains the final authority.

## Trust boundary

The current design intentionally splits model/planner output from trusted runtime state:

```text
        probabilistic / untrusted                         trusted control plane

┌─────────────────────────────────┐             ┌─────────────────────────────────┐
│ RemediationProposal             │             │ ExecutionContext                │
│                                 │             │                                 │
│ hypothesis                      │             │ authenticated actor             │
│ evidence references             │             │ granted authority               │
│ estimated_risk (advisory only)  │             │ human approval provenance       │
│ typed semantic action           │             │ operator override               │
│ preconditions                   │             │ policy version                  │
│ verification criteria           │             │ environment                     │
│ typed rollback action           │             │                                 │
└────────────────┬────────────────┘             └────────────────┬────────────────┘
                 │                                               │
                 └──────────────────────┬────────────────────────┘
                                        ▼
                         Deterministic Risk Classifier
                                        │
                               effective_risk
                                        ▼
                              Remediation Guard
                                        │
                              allow / deny / escalate
                                        ▼
                                Semantic Executor
```

**The proposal cannot approve itself, grant itself authority, disable an operator override, or choose the risk used for authorization.**

## Evidence contract

Evidence is data with provenance, not free-form context pasted into a prompt.

Each evidence reference carries an ID, kind, source URI, source system, collector identity, subject, observation/collection timestamps, optional observation window, optional freshness deadline, source trust metadata, and an optional SHA-256 integrity digest.

Important distinctions:

- `summary` is **untrusted evidence content**, never an instruction;
- `trust` describes the source/provenance class, not how strongly a model should believe a hypothesis;
- evidence weighting/ranking belongs to diagnosis logic, not to the evidence object itself;
- `fresh`, `stale`, and `freshness unknown` are distinct states;
- supporting and contradicting relations are explicit rather than hidden inside generated prose;
- missing telemetry may be represented explicitly instead of silently disappearing from context.

A deterministic validator rejects malformed provenance before a mutable proposal can pass the remediation guard.

## Core principles

1. **Observe freely.** Read-only telemetry and evidence collection are broadly available.
2. **Preserve provenance.** Evidence must retain source, collector, time, subject, freshness semantics, and integrity metadata when available.
3. **Reason probabilistically.** ML/LLM components may express uncertainty, hypotheses, confidence, alternatives, and advisory risk estimates.
4. **Propose structurally.** AI outputs typed remediation proposals, never arbitrary shell commands.
5. **Classify risk deterministically.** Effective operational risk is derived from trusted semantic action rules, not accepted from model output.
6. **Authorize from trusted context.** Identity, authority, human approval, operator override, environment, and policy version are supplied outside the proposal.
7. **Execute minimally.** Prefer the smallest reversible intervention that can validate the hypothesis.
8. **Verify independently.** Success is determined by measurable postconditions/SLO signals, not by the model that proposed the action.
9. **Rollback exactly.** A typed safe escape path is required before execution for action classes that demand it.
10. **Fail closed on ambiguity.** Missing context, malformed evidence/actions, risk under-reporting, or unknown actions reduce authority rather than expand it.

## Non-goals

This project is **not** intended to be:

- an unrestricted autonomous `kubectl`/SSH agent;
- an LLM wrapper around shell execution;
- a system where model confidence overrides policy;
- a system that trusts an AI-provided `human_approved: true` flag;
- a system that accepts model-declared risk as authorization truth;
- a system that treats retrieved text as trusted instructions;
- a mechanism for silently expanding its own production permissions;
- an opaque “AI fixed it” black box.

## High-level architecture

```text
Logs / Metrics / Traces / Events / Git / Runbooks / Incidents
                           │
                           ▼
                    Evidence Layer
               provenance / freshness / relations
                           │
                           ▼
              Diagnosis / Retrieval / ML / LLM
                           │
                    typed proposal
                           ▼
              Deterministic Risk Classifier
                           │
                    effective risk
                           ▼
              ┌─────────────────────────┐
              │   Remediation Guard     │◄──── trusted ExecutionContext
              │                         │
              │ evidence validation     │
              │ schema validation       │
              │ authority               │
              │ effective risk          │
              │ blast radius            │
              │ allow/deny rules        │
              │ preconditions           │
              │ rollback requirement    │
              │ operator override       │
              │ approval provenance     │
              │ policy provenance       │
              └────────────┬────────────┘
                           │ approved
                           ▼
                  Semantic Executor
                           │
                   dry-run / canary
                           ▼
                    Production Action
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
         Independent                Verification
          Watchdog                     │
              │                         │
              └──── abort/rollback ─────┘
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

## Proposal, not code

AI/planner components may produce proposal data such as:

```yaml
proposal:
  id: rem_20260810_001
  incident_id: inc_20260810_payment_5xx
  estimated_risk: R1  # advisory; deliberately wrong in this example
  hypothesis:
    type: unhealthy_new_deployment
    confidence: 0.87
  evidence:
    - id: ev_deploy_193
      kind: change
      uri: k8s://payments/deployment/payment-api/revision/193
      source: kubernetes-apiserver
      collector: infraheal-k8s-observer/v0
      subject: k8s://payments/deployment/payment-api
      summary: deployment changed before the error-rate increase
      observed_at: 2026-08-10T04:00:00Z
      collected_at: 2026-08-10T04:00:01Z
      fresh_until: 2026-08-10T04:15:00Z
      trust: high
    - id: ev_error_rate
      kind: metric
      uri: prometheus://payment-api/error-rate?window=5m
      source: prometheus
      collector: infraheal-prometheus-adapter/v0
      subject: k8s://payments/deployment/payment-api
      summary: 5xx rose immediately after revision 193 rollout
      observed_at: 2026-08-10T04:12:00Z
      collected_at: 2026-08-10T04:12:02Z
      fresh_until: 2026-08-10T04:14:00Z
      trust: high
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
    - p99_latency < 500ms
```

The deterministic risk classifier derives `rollback_deployment` as **R2**. Because the proposal under-reported it as R1, the current guard fails closed rather than trusting the model estimate.

Trusted runtime state is supplied separately:

```yaml
execution_context:
  authority: L5_APPROVAL
  human_approval:
    approval_id: CHG-1234
    approved_by: operator@example.com
  operator_override: false
  policy_version: mvp-v2
  environment: staging
  actor_id: controlplane:staging
```

The executor maps an approved semantic action such as `rollback_deployment` to trusted implementation code. It does not execute model-generated shell text.

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

## Repository map

```text
cmd/controlplane/        Go control-plane entry point
cmd/controller/          observe-only Kubernetes controller
api/v1alpha1/            HealingPolicy Kubernetes API
internal/domain/         proposal/action/context/evidence domain models
internal/risk/           deterministic effective-risk classifier
internal/policy/         remediation guard
internal/executor/       semantic executor boundary
internal/controller/     observe-only reconciliation logic
docs/MANIFESTO.md        design principles
docs/ARCHITECTURE.md     architecture and control flow
docs/THREAT_MODEL.md     failure and abuse model
docs/REMEDIATION_POLICY.md policy semantics
docs/LEARNING_ROADMAP.md research + learning plan
config/                  CRD, generated RBAC, manager and examples
test/e2e/                restricted-RBAC Kind safety test
examples/                sample proposals and trusted contexts
```

## Current milestones

Completed foundations include:

- closed typed remediation action schema;
- trusted execution context separated from planner output;
- deterministic effective-risk classification;
- fail-closed behavior for malformed or under-reported input;
- provenance-bearing evidence model and validator;
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

Early research and engineering scaffold. The repository is deliberately safety-first: autonomy will be added only after policy, audit, simulation, watchdog, rollback, evidence provenance, and verification mechanisms are independently testable.

## License

Apache-2.0.
