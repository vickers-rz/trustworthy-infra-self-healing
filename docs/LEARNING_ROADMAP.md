# Learning & Research Roadmap

This repository is both a research project and a deliberate path from cloud operations toward infrastructure/platform engineering with trustworthy AI-assisted automation.

## Phase 0 — Foundations and safety thesis

**Learn:** Go interfaces/errors/context/testing; SRE control loops; failure domains; least privilege; idempotency; rollback semantics.

**Build:** typed proposal model, deterministic guard, decision reasons, mock executor, unit tests.

**Exit criterion:** no mutation path exists outside typed semantic actions and fail-closed policy.

## Phase 1 — Go control plane

**Learn:** concurrency, channels, worker pools, cancellation, structured logging, HTTP APIs, configuration, fuzzing and race testing.

**Build:** proposal API, policy engine, append-only audit records, action registry, dry-run executor and fuzz tests against guard bypasses.

## Phase 2 — Observability and evidence graph

**Learn:** OpenTelemetry, Prometheus, logs/traces/metrics correlation, Kubernetes events, deployment metadata and SLOs.

**Build:** evidence adapters, canonical evidence references, freshness/trust metadata, incident timeline and dependency context.

**Exit criterion:** every hypothesis and remediation can cite a reproducible evidence pack.

## Phase 3 — Safe remediation lab

**Learn:** Kubernetes controllers/operators, reconciliation, GitOps, rollout/rollback, canarying and chaos engineering.

**Build:** disposable local cluster lab; restart/rollback/scale semantic executors; precondition checks; simulation/dry-run; independent watchdog; exact rollback tests.

## Phase 4 — LLM/RAG diagnosis, not authority

**Learn:** retrieval quality, prompt injection, structured outputs, uncertainty calibration, tool boundaries and evaluation.

**Build:** RAG over runbooks/postmortems/Git history; evidence-backed RCA hypotheses; contradictory-evidence handling; schema-constrained proposal generation.

**Evaluate:** diagnosis accuracy, evidence precision/recall, unsupported-claim rate, proposal validity and policy-denial correctness.

## Phase 5 — ML for detection/ranking

**Learn:** anomaly/change-point detection, time-series features, incident similarity, ranking, calibration and drift.

**Build:** anomaly detector, root-cause ranker, remediation-effectiveness model and confidence calibration.

**Constraint:** learned output changes belief/ranking only; it cannot grant authority or mutate hard policy.

## Phase 6 — Trustworthiness research

Build reproducible experiments around stale signals, RAG poisoning, prompt injection, contradictory evidence, controller crash, watchdog partition, rollback failure, human override and blast-radius misclassification.

Publish scenarios, benchmark datasets, ADRs and postmortems of the project's own failures.

## Portfolio outcome

The repository should ultimately demonstrate production-style Go control-plane engineering, Kubernetes/SRE competence, deterministic safety policy, observability/evidence provenance, LLM/RAG engineering with explicit trust boundaries, ML for detection/ranking, and chaos/rollback/human-override semantics.
