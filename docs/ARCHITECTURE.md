# Architecture

## Control loop

```text
OBSERVE → DIAGNOSE → PROPOSE → AUTHORIZE → SIMULATE → CANARY → EXECUTE → VERIFY → LEARN
                                │                                  │
                                └──────── deny / escalate          └── watchdog → abort/rollback
```

The architecture intentionally separates components that *form beliefs* from components that *hold authority*.

## Evidence layer

Inputs may include metrics, logs, traces, Kubernetes events, cloud control-plane events, Git history, deployment revisions, runbooks, postmortems, dependency topology, and operator annotations. Every item is represented by a provenance-bearing reference.

RAG belongs here: it is an **evidence acquisition mechanism**, not an actuator.

## Diagnosis layer

Rules, statistical detectors, classical ML, and LLMs may jointly produce hypotheses and rank likely root causes. Outputs are uncertain by design and must preserve confidence and supporting/contradicting evidence.

## Proposal layer

A planner converts a hypothesis into a typed `Proposal`. The schema describes semantic action, target, risk class, preconditions, rollback, verification criteria, evidence and policy version. Arbitrary shell is not part of the schema.

## Remediation Guard

The guard is deterministic and fail-closed. It evaluates:

- action allowlist and schema validity;
- remediation authority;
- declared and catalogued risk;
- blast radius;
- human/operator override;
- preconditions;
- rollback requirement;
- verification requirement;
- explicit approval where required;
- policy version and provenance.

Model confidence cannot override a denied policy decision.

## Semantic executor

Approved actions map to trusted implementations such as `restart_workload`, `rollback_deployment` or `scale_workload`. Parameters are validated; the executor never accepts generated command text as an execution primitive.

## Independent watchdog

The watchdog owns deadlines, abort signals and safety invariants. A future production design should deploy it in a distinct failure domain from the planner/controller so LLM outage, RAG failure, controller crash or network partition cannot remove the escape path.

## Verification

Verification uses externally observable postconditions: SLOs, readiness, error rate, latency, saturation, dependency health and application-specific invariants. The proposing model does not self-certify success.

## Learning

Outcomes may update incident similarity, remediation ranking, confidence calibration and hypothesis priors. They may not directly rewrite hard policy or increase authority. Policy evolution is an explicit reviewed artifact, ideally GitOps-managed.

## Initial package boundaries

```text
internal/domain     typed contracts
internal/policy     deterministic authorization
internal/executor   semantic action boundary
```

Future packages will add `evidence`, `diagnosis`, `audit`, `watchdog`, `simulation`, `verification`, and provider adapters.
