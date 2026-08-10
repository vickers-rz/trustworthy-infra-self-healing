# Threat Model

## Assets

The system protects production availability, integrity, confidentiality, operator authority, evidence provenance, auditability, and the correctness/replayability of remediation decisions.

## Primary threats

| Threat | Failure mode | Required mitigation |
|---|---|---|
| LLM hallucination | invents cause or unsafe fix | typed proposal, bounded evidence references, deterministic guard |
| Prompt injection in logs/runbooks | attacker text asks model to execute | retrieved text is untrusted evidence; no direct execution authority |
| Command injection | model smuggles shell through parameters | closed typed semantic actions; no generic command/parameter primitive |
| Risk under-reporting | model labels R2/R3 action as R1 to gain delegation | deterministic effective-risk classifier; under-report mismatch fails closed |
| Forged approval | model emits `human_approved: true` | approval exists only in trusted ExecutionContext with provenance |
| Forged authority | model requests L4/L5 for itself | authority is supplied by trusted runtime, never Proposal |
| Operator override spoofing | model clears or ignores an operator stop | override lives in trusted runtime and has absolute precedence |
| Forged evidence provenance | model embeds invented `source`, `collector`, `trust`, or timestamps | Proposal contains only trusted bundle/item IDs; EvidenceBundle is supplied independently |
| Evidence-bundle substitution | model references a different/invented bundle | proposal bundle ID must exactly match trusted bundle ID and incident ID |
| Evidence-reference smuggling | model references nonexistent or duplicate item IDs | Guard resolves selected IDs inside trusted bundle and fails closed |
| Stale-evidence replay | old telemetry is replayed as current justification | explicit trusted DecisionTime + deterministic freshness policy |
| Unknown freshness treated as current | missing TTL silently behaves as fresh | freshness is three-state; unknown operational freshness fails closed |
| Reference-only authorization | runbook/postmortem is treated as proof of current state | mutable action requires at least one fresh operational evidence item |
| Privilege escalation | controller exceeds intended scope | least privilege/RBAC, action-scoped credentials |
| Autonomous policy drift | repeated success expands permissions | learned state cannot mutate hard policy; policy changes via review/GitOps |
| Excessive blast radius | correct repair applied too broadly | explicit target scope, deterministic risk, canary/minimum perturbation |
| Human/AI conflict | automation fights operator | operator override has absolute precedence |
| Correlated controller failure | planner failure disables rollback | independent watchdog and rollback path |
| Evidence poisoning | manipulated telemetry/RAG biases diagnosis | independent provenance, source registry, contradictory evidence, multi-signal checks |
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
- Proposal cannot embed trusted evidence provenance; it can only reference bundle/item IDs.
- Trusted EvidenceBundle ID and incident ID must match Proposal references.
- Unknown or duplicate Proposal evidence IDs fail closed.
- DecisionTime is explicit trusted runtime state.
- Stale or freshness-unknown selected operational evidence fails closed.
- Reference evidence cannot alone authorize mutation.
- Operator override is trusted runtime state and always cancels automation ownership.
- R4/R5 mutation is denied.
- R3 requires L5 authority and explicit human approval provenance.
- R2 is not autonomously delegated in the initial policy.
- R2+ requires an explicit typed rollback/compensation action.
- Mutable actions require preconditions and independent verification criteria.

## Residual evidence risks

The current bundle separation is an architectural trust boundary, not yet a cryptographic proof boundary.

Still to harden:

- authenticate collectors that create EvidenceBundle items;
- assign source trust from a trusted registry instead of trusting incoming labels;
- seal/sign bundle-level content and bind it into audit records;
- prevent a compromised Evidence Plane from issuing forged-but-well-formed evidence;
- add independent corroboration requirements for higher-risk remediation;
- bind diagnosis relations to a versioned diagnosis artifact.

The key property already enforced is narrower but important: **the probabilistic Proposal itself cannot fabricate provenance and have the Guard treat that fabricated object as trusted evidence.**

## Abuse-resistant learning

Model output, pseudo-labels, historical correlations, and remediation outcome models are advisory state. They can change ranking or confidence, but they cannot:

- grant authority;
- lower deterministic risk;
- weaken a deny rule;
- remove approval requirements;
- clear an operator override;
- rewrite trusted evidence provenance;
- rewrite trusted decision time;
- promote themselves to ground truth through self-reinforcing training loops.

## Future trust-boundary work

1. trusted source/collector registry;
2. signed/sealed EvidenceBundle and approval artifacts;
3. policy/classifier/bundle hashes in audit records;
4. action-scoped execution capabilities rather than broad credentials;
5. separate failure domain for watchdog/rollback;
6. blast-radius classifier using topology and service criticality;
7. executor API that cannot be called without an authorization decision token;
8. corroboration/diversity requirements for R2/R3 actions.
