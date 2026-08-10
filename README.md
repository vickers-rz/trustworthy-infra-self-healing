# Trustworthy Infra Self-Healing

> **Reason probabilistically. Authorize deterministically.**

A research-oriented, Go-first control plane for **trustworthy, transparent, explainable infrastructure self-healing**.

The project explores a simple but demanding question:

> How can probabilistic AI safely operate deterministic production infrastructure?

The answer is not “give an LLM shell access.” This repository separates **observation, diagnosis, proposal, authorization, execution, verification, rollback, and learning** into explicit layers with independent safety boundaries.

## Design thesis

This project inherits a core idea from HomeMind: **learning authority is not actuation authority**.

For infrastructure, that becomes:

> **Diagnostic Authority ≠ Remediation Authority**

An AI component may read logs, metrics, traces, deployment history, runbooks, incident reports, and topology; generate hypotheses; retrieve evidence with RAG; rank candidate causes; and produce a structured remediation proposal. None of those capabilities imply permission to mutate production state.

The deterministic control plane remains the final authority.

## Core principles

1. **Observe freely.** Read-only telemetry and evidence collection are broadly available.
2. **Reason probabilistically.** ML/LLM components may express uncertainty, hypotheses, confidence, and alternatives.
3. **Propose structurally.** AI outputs typed remediation proposals, never arbitrary shell commands.
4. **Authorize deterministically.** Policy, risk, blast radius, preconditions, identity, and rollback requirements decide whether an action may proceed.
5. **Execute minimally.** Prefer the smallest reversible intervention that can validate the hypothesis.
6. **Verify independently.** Success is determined by measurable postconditions/SLO signals, not by the model that proposed the action.
7. **Rollback exactly.** A safe escape path is required before execution for any mutable action class that demands it.
8. **Learn with provenance.** Every conclusion and action must retain evidence, model, policy, approval, execution, and outcome lineage.

## Non-goals

This project is **not** intended to be:

- an unrestricted autonomous `kubectl`/SSH agent;
- an LLM wrapper around shell execution;
- a system where model confidence overrides policy;
- a mechanism for silently expanding its own production permissions;
- an opaque “AI fixed it” black box.

## High-level architecture

```text
Logs / Metrics / Traces / Events / Git / Runbooks / Incidents
                           │
                           ▼
                    Evidence Layer
                           │
                           ▼
              Diagnosis / Retrieval / ML / LLM
                           │
                    structured proposal
                           ▼
              ┌─────────────────────────┐
              │   Remediation Guard     │
              │                         │
              │ schema validation       │
              │ authority level         │
              │ operational risk        │
              │ blast radius            │
              │ allow/deny rules        │
              │ preconditions           │
              │ rollback requirement    │
              │ operator override       │
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

### Remediation authority

| Level | Capability |
|---|---|
| `L0 OBSERVE` | read telemetry and state |
| `L1 DIAGNOSE` | produce root-cause hypotheses |
| `L2 PROPOSE` | generate structured remediation proposals |
| `L3 PREPARE` | render a bounded execution plan without applying it |
| `L4 DELEGATED` | execute pre-authorized low-risk playbooks |
| `L5 APPROVAL` | execute only after explicit human approval |
| `DENY` | action is never autonomously executable |

### Operational risk / blast radius

| Class | Meaning | Example |
|---|---|---|
| `R0` | observation only | query metrics/logs |
| `R1` | ephemeral/single workload | restart one stateless pod |
| `R2` | service-local reversible change | rollback one deployment |
| `R3` | multi-service/stateful impact | database failover |
| `R4` | business/security critical | IAM/network/schema mutation |
| `R5` | catastrophic or irreversible | destructive storage/database action |

Authority and risk are evaluated together. A high delegation level does **not** imply permission for a high-risk action.

## Proposal, not code

AI components produce data such as:

```yaml
proposal:
  id: rem_20260810_001
  incident:
    service: payment-api
    symptom: elevated_5xx
  hypothesis:
    type: unhealthy_new_deployment
    confidence: 0.87
  evidence:
    - deployment_changed_12m_ago
    - error_rate_increased_after_change
    - previous_revision_was_healthy
  action:
    type: rollback_deployment
    target: payment-api
    from_revision: "193"
    to_revision: "192"
  risk_class: R2
  rollback:
    strategy: restore_revision
    revision: "193"
  verification:
    - error_rate_5m < 1%
    - p99_latency < 500ms
```

The executor maps an approved semantic action such as `rollback_deployment` to trusted implementation code. It does not execute model-generated shell text.

## Repository map

```text
cmd/controlplane/        Go control-plane entry point
internal/domain/         typed proposal/risk/authority models
internal/policy/         deterministic remediation guard
internal/executor/       semantic executor boundary
docs/MANIFESTO.md        design principles
docs/ARCHITECTURE.md     architecture and control flow
docs/THREAT_MODEL.md     failure and abuse model
docs/REMEDIATION_POLICY.md policy semantics
docs/LEARNING_ROADMAP.md research + learning plan
config/                  example policies
examples/                sample proposals and incidents
```

## Initial milestone

The first milestone intentionally avoids live production mutation. It focuses on the part that must be correct before autonomy:

- typed remediation proposal schema;
- authority and risk classification;
- deterministic policy decisions;
- fail-closed behavior for malformed input;
- mandatory rollback where required;
- explicit rule-hit explanations;
- audit/provenance-ready decision records;
- mock semantic executor;
- unit/fuzz tests for guard bypass attempts.

## Development

```bash
go test ./...
go run ./cmd/controlplane
```

## Project status

Early research and architecture scaffold. The repository is deliberately safety-first: autonomy will be added only after policy, audit, simulation, watchdog, and rollback mechanisms are independently testable.

## License

Apache-2.0.
