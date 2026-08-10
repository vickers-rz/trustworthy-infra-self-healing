# Threat Model

## Assets

The system protects production availability, integrity, confidentiality, operator authority, auditability, and the correctness of remediation decisions.

## Primary threats

| Threat | Failure mode | Required mitigation |
|---|---|---|
| LLM hallucination | invents cause or unsafe fix | typed proposal, evidence citations, deterministic guard |
| Prompt injection in logs/runbooks | attacker text asks model to execute | retrieved text is untrusted evidence; no direct execution authority |
| Command injection | model smuggles shell through parameters | closed typed semantic actions; no generic command/parameter primitive |
| Risk under-reporting | model labels R2/R3 action as R1 to gain delegation | deterministic effective-risk classifier; under-report mismatch fails closed |
| Forged approval | model emits `human_approved: true` | approval exists only in trusted ExecutionContext with provenance |
| Forged authority | model requests L4/L5 for itself | authority is supplied by trusted runtime, never Proposal |
| Operator override spoofing | model clears or ignores an operator stop | override lives in trusted runtime and has absolute precedence |
| Privilege escalation | controller exceeds intended scope | least privilege/RBAC, action-scoped credentials |
| Autonomous policy drift | repeated success expands permissions | learned state cannot mutate hard policy; policy changes via review/GitOps |
| Stale telemetry | repair based on obsolete state | freshness constraints and precondition re-check immediately before execution |
| Excessive blast radius | correct repair applied too broadly | explicit target scope, deterministic risk, canary/minimum perturbation |
| Human/AI conflict | automation fights operator | operator override has absolute precedence |
| Correlated controller failure | planner failure disables rollback | independent watchdog and rollback path |
| Evidence poisoning | manipulated telemetry/RAG biases diagnosis | provenance, source trust, contradictory evidence, multi-signal checks |
| Direct executor bypass | code calls mutation adapter without policy | future executor must require authorization artifact/capability, plus RBAC separation |

## Hard MVP invariants

- No arbitrary shell execution primitive.
- No generic `map[string]string` execution parameter channel.
- Unknown or malformed semantic actions fail closed.
- Proposal-declared risk is advisory only.
- Deterministic `effective_risk` is the authorization source of truth.
- Risk under-reporting fails closed.
- Proposal cannot grant itself authority.
- Proposal cannot assert human approval.
- Operator override is trusted runtime state and always cancels automation ownership.
- R4/R5 mutation is denied.
- R3 requires L5 authority and explicit human approval provenance.
- R2 is not autonomously delegated in the initial policy.
- R2+ requires an explicit typed rollback/compensation action.
- Mutable actions require preconditions and independent verification criteria.
- Missing evidence provenance denies execution.

## Abuse-resistant learning

Model output, pseudo-labels, historical correlations, and remediation outcome models are advisory state. They can change ranking or confidence, but they cannot:

- grant authority;
- lower deterministic risk;
- weaken a deny rule;
- remove approval requirements;
- clear an operator override;
- promote themselves to ground truth through self-reinforcing training loops.

## Future trust-boundary work

The following are explicit future hardening tasks:

1. signed/immutable approval artifacts;
2. policy and classifier version hashes in audit records;
3. action-scoped execution capabilities rather than broad credentials;
4. separate failure domain for watchdog/rollback;
5. freshness and integrity metadata for Evidence objects;
6. blast-radius classifier using topology and service criticality;
7. executor API that cannot be called without an authorization decision token.
