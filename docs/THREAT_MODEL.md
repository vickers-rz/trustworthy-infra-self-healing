# Threat Model

## Assets

The system protects production availability, integrity, confidentiality, operator authority, auditability and the correctness of remediation decisions.

## Primary threats

| Threat | Failure mode | Required mitigation |
|---|---|---|
| LLM hallucination | invents cause or unsafe fix | typed proposal, evidence citations, deterministic guard |
| Prompt injection in logs/runbooks | attacker text asks model to execute | retrieved text is untrusted evidence; no direct execution authority |
| Command injection | model smuggles shell through parameters | semantic allowlist, strict schemas, no generic shell primitive |
| Privilege escalation | controller exceeds intended scope | least privilege/RBAC, action-scoped credentials |
| Autonomous policy drift | repeated success expands permissions | learned state cannot mutate hard policy; policy changes via review/GitOps |
| Stale telemetry | repair based on obsolete state | freshness constraints and precondition re-check immediately before execution |
| Excessive blast radius | correct repair applied too broadly | explicit target scope, risk matrix, canary/minimum perturbation |
| Human/AI conflict | automation fights operator | operator override has absolute precedence |
| Correlated controller failure | planner failure disables rollback | independent watchdog and rollback path |
| Evidence poisoning | manipulated telemetry/RAG biases diagnosis | provenance, source trust, contradictory evidence, multi-signal checks |

## Hard MVP invariants

- No arbitrary shell execution primitive.
- Unknown semantic actions fail closed.
- R4/R5 mutation is denied.
- R3 requires L5 authority and explicit human approval.
- R2 is not autonomously delegated in the initial policy.
- Operator override always cancels automation ownership.
- R2+ requires an explicit rollback plan.
- Mutable actions require preconditions and independent verification criteria.
- Missing evidence provenance denies execution.

## Abuse-resistant learning

Model output, pseudo-labels and historical correlations are advisory state. They can change ranking or confidence, but they cannot grant authority, weaken a deny rule, remove approval, or promote themselves to ground truth through self-reinforcing training loops.
