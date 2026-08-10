# Remediation Policy

A remediation decision is evaluated as **authority × operational risk**, plus hard invariants.

## Authority

- `L0_OBSERVE`: read-only state
- `L1_DIAGNOSE`: generate hypotheses
- `L2_PROPOSE`: create structured remediation proposal
- `L3_PREPARE`: render bounded plan/dry-run
- `L4_DELEGATED`: execute pre-authorized low-risk playbook
- `L5_APPROVAL`: execute only with explicit human approval
- `DENY`: never executable

## Risk

- `R0`: observation only
- `R1`: ephemeral/single-workload action
- `R2`: service-local reversible change
- `R3`: stateful or multi-service impact
- `R4`: business/security-critical mutation
- `R5`: catastrophic or irreversible action

## MVP matrix

| Risk | L4 delegated | L5 + explicit approval |
|---|---:|---:|
| R0 | allow | allow |
| R1 | allow if all invariants pass | allow |
| R2 | deny (MVP) | allow if rollback + verification pass |
| R3 | deny | allow only for catalogued actions with rollback |
| R4 | deny | deny (MVP) |
| R5 | deny | deny |

The matrix is deliberately conservative. Expanding it must be an explicit, reviewable policy change; it is never a learned side effect.

## Semantic action catalogue

| Action | Maximum risk | Notes |
|---|---:|---|
| `restart_workload` | R1 | stateless/single workload only |
| `rollback_deployment` | R2 | previous revision and compatibility required |
| `scale_workload` | R2 | bounded min/max and capacity checks |
| `drain_node` | R3 | approval-gated |
| `failover_database` | R3 | approval-gated; provider-specific invariants required |

There is deliberately no `run_shell`, `kubectl`, `ssh`, or free-form command action.

## Decision record

Every decision should eventually persist proposal/incident IDs, policy version/hash, identity, authority, risk/blast radius, matched rules, evidence, approval provenance, precondition results, execution result, verification result, rollback result, and operator override events.
