# Trustworthy Infra Self-Healing

> **Reason probabilistically. Authorize deterministically.**

A research-oriented, Go-first control plane for **trustworthy, transparent, explainable infrastructure self-healing**.

The project explores a simple but demanding question:

> How can probabilistic AI safely assist deterministic production infrastructure without becoming an unrestricted production operator?

The answer is not “give an LLM shell access.” This repository separates **observation, diagnosis, proposal, trusted execution context, risk classification, authorization, execution, verification, rollback, and learning** into explicit layers with independent safety boundaries.

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

## Core principles

1. **Observe freely.** Read-only telemetry and evidence collection are broadly available.
2. **Reason probabilistically.** ML/LLM components may express uncertainty, hypotheses, confidence, alternatives, and advisory risk estimates.
3. **Propose structurally.** AI outputs typed remediation proposals, never arbitrary shell commands.
4. **Classify risk deterministically.** Effective operational risk is derived from trusted semantic action rules, not accepted from model output.
5. **Authorize from trusted context.** Identity, authority, human approval, operator override, environment, and policy version are supplied outside the proposal.
6. **Execute minimally.** Prefer the smallest reversible intervention that can validate the hypothesis.
7. **Verify independently.** Success is determined by measurable postconditions/SLO signals, not by the model that proposed the action.
8. **Rollback exactly.** A typed safe escape path is required before execution for action classes that demand it.
9. **Learn with provenance.** Every conclusion and action retains evidence, model, policy, approval, execution, and outcome lineage.
10. **Fail closed on ambiguity.** Missing context, malformed typed actions, risk under-reporting, or unknown actions reduce authority rather than expand it.

## Non-goals

This project is **not** intended to be:

- an unrestricted autonomous `kubectl`/SSH agent;
- an LLM wrapper around shell execution;
- a system where model confidence overrides policy;
- a system that trusts an AI-provided `human_approved: true` flag;
- a system that accepts model-declared risk as authorization truth;
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
                    typed proposal
                           ▼
              Deterministic Risk Classifier
                           │
                    effective risk
                           ▼
              ┌─────────────────────────┐
              │   Remediation Guard     │◄──── trusted ExecutionContext
              │                         │
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
    - uri: deploy://payment-api/revision/193
      summary: deployment changed before the error-rate increase
    - uri: metric://prometheus/payment-api/error-rate
      summary: 5xx rose immediately after revision 193 rollout
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
internal/domain/         typed proposal/action/trusted-context models
internal/risk/           deterministic effective-risk classifier
internal/policy/         remediation guard
internal/executor/       semantic executor boundary
docs/MANIFESTO.md        design principles
docs/ARCHITECTURE.md     architecture and control flow
docs/THREAT_MODEL.md     failure and abuse model
docs/REMEDIATION_POLICY.md policy semantics
docs/LEARNING_ROADMAP.md research + learning plan
config/                  example policies
examples/                sample proposals and trusted contexts
```

## Initial milestone

The first milestone intentionally avoids live production mutation. It focuses on the part that must be correct before autonomy:

- closed typed remediation action schema;
- trusted execution context separated from planner output;
- deterministic effective-risk classification;
- authority and risk policy;
- fail-closed behavior for malformed or under-reported input;
- mandatory typed rollback where required;
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

Early research and architecture scaffold. The repository is deliberately safety-first: autonomy will be added only after policy, audit, simulation, watchdog, rollback, evidence provenance, and verification mechanisms are independently testable.

## License

Apache-2.0.
