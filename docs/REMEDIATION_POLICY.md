# Remediation Policy

A remediation decision is evaluated as **trusted execution authority × deterministic operational risk**, plus hard invariants.

## Trust boundary

The planner/model may propose:

- a hypothesis;
- evidence;
- a typed action;
- an advisory `estimated_risk`;
- preconditions;
- verification criteria;
- a typed rollback action.

The planner/model may **not** authoritatively provide:

- execution authority;
- human approval;
- operator override state;
- policy version;
- authenticated actor identity;
- effective operational risk.

Those values come from trusted control-plane context and deterministic classifiers.

## Authority

- `L0_OBSERVE`: read-only state
- `L1_DIAGNOSE`: generate hypotheses
- `L2_PROPOSE`: create structured remediation proposal
- `L3_PREPARE`: render bounded plan/dry-run
- `L4_DELEGATED`: execute pre-authorized low-risk playbook
- `L5_APPROVAL`: execute only with explicit human approval provenance
- `DENY`: never executable

`L5_APPROVAL` is not “more autonomous” than `L4_DELEGATED`; it is a different execution authority mode requiring a human authorization artifact.

## Risk

- `R0`: observation only
- `R1`: ephemeral/single-workload action
- `R2`: service-local reversible change
- `R3`: stateful or multi-service impact
- `R4`: business/security-critical mutation
- `R5`: catastrophic or irreversible action

## Risk source of truth

`Proposal.estimated_risk` is advisory.

The authorization source of truth is `effective_risk`, derived by the deterministic risk engine from semantic action type and, in later phases, target scope, environment, criticality, statefulness, dependency topology, and blast radius.

If `estimated_risk < effective_risk`, the MVP fails closed. A model cannot gain privilege by classifying its own action as safer than the control plane does.

## MVP matrix

| Effective risk | L4 delegated | L5 + explicit approval |
|---|---:|---:|
| R0 | allow | allow |
| R1 | allow if all invariants pass | allow with approval provenance |
| R2 | deny (MVP) | allow if typed rollback + verification pass |
| R3 | deny | allow only for catalogued actions with typed rollback |
| R4 | deny | deny (MVP) |
| R5 | deny | deny |

The matrix is deliberately conservative. Expanding it must be an explicit, reviewable policy change; it is never a learned side effect.

## Semantic action catalogue

| Action | Base effective risk | Notes |
|---|---:|---|
| `restart_workload` | R1 | stateless/single workload only |
| `rollback_deployment` | R2 | previous revision and compatibility required |
| `scale_workload` | R2 | bounded min/max and capacity checks |
| `uncordon_node` | R2 | compensation/recovery action; future scope checks required |
| `drain_node` | R3 | approval-gated |
| `failover_database` | R3 | approval-gated; provider-specific invariants required |

There is deliberately no `run_shell`, `kubectl`, `ssh`, or free-form command action.

## Typed action invariant

Actions use closed, typed payloads. The executor must reject:

- an unknown action type;
- multiple payload variants in one action;
- a payload that does not match the declared type;
- invalid typed fields;
- arbitrary command strings hidden inside generic parameter maps.

Rollback/compensation is also a typed semantic action.

## Human approval provenance

Approval is represented as trusted runtime metadata such as:

```yaml
human_approval:
  approval_id: CHG-1234
  approved_by: operator@example.com
```

A boolean inside a model-generated proposal is never sufficient evidence of approval.

## Decision record

Every decision should eventually persist:

- proposal/incident IDs;
- deterministic effective risk and classifier version;
- planner estimated risk and any disagreement;
- policy version/hash;
- actor identity;
- authority;
- risk/blast radius;
- matched rules;
- evidence;
- approval provenance;
- operator override state;
- precondition results;
- execution result;
- verification result;
- rollback result.
