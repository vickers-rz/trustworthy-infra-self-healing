# Learning & Research Roadmap

This repository is both a research project and a deliberate path from cloud operations toward **infrastructure/platform software engineering with trustworthy AI-assisted automation**.

The roadmap is milestone-driven rather than course-driven: every phase must leave behind runnable code, tests, failure experiments, and design artifacts.

The long-term research target is broader than automated repair: build a control plane that can make **bounded reliability decisions under uncertainty**, including observation, abstention, containment, escalation, rollback, and recovery-oriented actions.

## Phase 0 — Foundations and safety thesis

**Learn:** Go interfaces/errors/context/testing; SRE control loops; failure domains; least privilege; idempotency; rollback semantics.

**Build:** typed proposal/action model, trusted execution context, deterministic risk classifier, remediation guard, decision reasons, mock executor, unit tests.

**Exit criterion:** no model/planner field can directly grant execution authority, assert human approval, choose the authorization risk, or smuggle arbitrary command text into an action.

## Phase 1 — Go control plane

**Learn:** concurrency, channels, worker pools, cancellation, structured logging, HTTP APIs, configuration, fuzzing, profiling, race testing, retries/backoff, and graceful shutdown.

**Build:** proposal API, append-only audit records, typed action registry, dry-run executor, authorization artifact concept, and fuzz tests against guard bypasses.

**Exit criterion:** a long-running Go service can accept typed proposals and trusted runtime context, classify risk, deny unsafe requests, and produce reconstructable decision records.

## Phase 1.5 — Kubernetes programming early

This phase is intentionally moved forward because writing Kubernetes controllers is a key transition signal from “operating infrastructure” to “developing infrastructure software”.

**Learn:**

- Kubernetes API machinery;
- desired/current state reconciliation;
- CRDs;
- controller-runtime / Kubebuilder;
- status conditions;
- finalizers;
- owner references;
- watches/caches;
- RBAC;
- leader election;
- idempotent reconciliation;
- controller testing.

### Milestone 1 — Observe-only `HealingPolicy`

The first controller is intentionally incapable of remediation. It observes an `apps/v1 Deployment`, writes only the `HealingPolicy/status` subresource, and watches Deployment changes so observation is event-driven as well as periodic.

Progress:

- [x] define `infraheal.io/v1alpha1` and `HealingPolicy`;
- [x] add a namespaced CRD with a status subresource;
- [x] reconcile a Deployment target without mutating it;
- [x] report target resolution and observed availability as status conditions;
- [x] watch Deployment changes and enqueue matching policies;
- [x] use target-side RBAC containing only `get/list/watch`;
- [x] add fake-client tests proving Deployment spec remains unchanged;
- [x] pin a Go 1.24-compatible controller-runtime/Kubernetes dependency line;
- [x] commit and CI-check the Go module graph;
- [ ] add a reproducible Kind integration test;
- [ ] generate CRD/deepcopy/RBAC artifacts with controller-gen instead of maintaining the bootstrap deepcopy file manually;
- [ ] add `HealingRun` as a separate execution/audit lifecycle resource after the observation contract is stable.

**Exit criterion:** controller restarts are safe, reconciliation is idempotent, status is meaningful, the Kind test is reproducible, target-side RBAC remains read-only, and no LLM is involved yet.

## Phase 2 — Observability, evidence graph, and failure-domain context

**Learn:** OpenTelemetry, Prometheus, logs/traces/metrics correlation, Kubernetes events, deployment metadata, SLOs, provenance, telemetry freshness, dependency topology, failure domains, blast radius, and fault-propagation models.

**Build:** evidence adapters, canonical evidence references, freshness/trust metadata, incident timeline, supporting/contradicting links, dependency context, explicit failure-domain labels, and propagation-edge representations.

The evidence plane should be able to answer not only “what is broken?” but also:

- where the failure appears to originate;
- which failure domains are currently affected;
- which clean domains are at risk;
- which dependency/change/replication paths may propagate bad state;
- whether the observed blast radius is expanding or stable.

**Exit criterion:** every hypothesis and remediation can cite a reproducible EvidenceBundle; missing or stale evidence is explicit rather than silently ignored; affected and at-risk failure domains can be reconstructed from trusted evidence.

## Phase 3 — Safe remediation and containment lab

**Learn:** reconciliation, GitOps, rollout/rollback, canarying, chaos engineering, failure injection, verification design, containment strategy, recovery boundaries, and blast-radius control.

**Build:** disposable local cluster lab; restart/rollback/scale semantic executors; precondition checks; simulation/dry-run; independent watchdog; typed compensation; exact rollback tests; failure-domain-aware containment proposals and policies.

Initial incidents:

1. deployment regression;
2. CrashLoopBackOff;
3. OOMKilled;
4. readiness failure;
5. dependency latency;
6. CPU saturation;
7. bad rollout propagating across multiple workloads;
8. dependency failure whose impact expands across service boundaries.

For the propagation scenarios, compare at least three outcomes:

- immediate repair attempt;
- containment-first strategy;
- abstain/escalate because evidence is insufficient.

Measure blast radius, time to containment, recovery time, rollback success, and unintended side effects.

**Exit criterion:** failures can be injected and recovered reproducibly without AI diagnosis; containment can be evaluated independently of repair; deterministic policy can reject an action that would expand the failure domain.

## Phase 4 — LLM/RAG diagnosis, not authority

**Learn:** retrieval quality, hybrid retrieval, temporal relevance, prompt injection, structured outputs, uncertainty calibration, tool boundaries, citations, abstention, and evaluation.

**Build:** RAG over runbooks/postmortems/Git history; evidence-backed RCA hypotheses; explicit supporting and contradictory evidence; schema-constrained proposal generation; propagation hypotheses that identify likely origin, affected domains, at-risk domains, and candidate containment points.

**Evaluate:** diagnosis accuracy, evidence precision/recall, unsupported-claim rate, contradiction handling, proposal validity, risk under-reporting rate, policy-denial correctness, propagation-hypothesis accuracy, and inappropriate-action rate.

**Invariant:** RAG/LLM output can change hypotheses and proposals, never trusted execution context or hard policy.

The LLM must be allowed to return “insufficient evidence”, “contain before repair”, or “escalate to operator” as legitimate outcomes. The benchmark should not reward automation for acting when abstention is safer.

## Phase 5 — ML for detection/ranking

**Learn:** anomaly/change-point detection, time-series features, incident similarity, ranking, calibration, drift, and class imbalance.

**Build:** anomaly detector, root-cause ranker, incident-similarity model, remediation-effectiveness model, propagation-risk ranker, and confidence calibration.

**Constraint:** learned output changes belief/ranking only; it cannot grant authority, lower deterministic effective risk, clear an override, or mutate hard policy.

## Phase 6 — Trustworthiness and reliability-decision research

Build reproducible experiments around:

- stale signals;
- telemetry disagreement;
- RAG poisoning;
- prompt injection;
- risk under-reporting;
- forged approval attempts;
- malformed typed actions;
- controller crash;
- watchdog partition;
- rollback failure;
- human override;
- blast-radius misclassification;
- direct executor bypass attempts;
- model/vector-store/policy-engine outage;
- propagation-path misidentification;
- delayed containment;
- containment that accidentally expands impact;
- contradictory evidence about affected failure domains;
- safe abstention versus premature automated action;
- escalation correctness for high-risk or ambiguous incidents.

### Reliability decision benchmark

Add a benchmark where the system must choose among distinct classes of next action:

1. `OBSERVE_MORE` — gather additional evidence;
2. `ABSTAIN` — do not mutate because confidence/evidence quality is insufficient;
3. `CONTAIN` — prevent further fault propagation;
4. `PREPARE_RECOVERY` — render rollback/failover/recovery steps without execution;
5. `EXECUTE_LOW_RISK` — apply a pre-authorized bounded action;
6. `ESCALATE` — require explicit operator decision;
7. `DENY` — reject unsafe or policy-prohibited action.

Evaluate not only whether the incident is eventually repaired, but whether the chosen decision minimizes expected operational harm under uncertainty.

Publish scenarios, benchmark datasets, ADRs, audit traces, and postmortems of the project's own failures.

## Portfolio outcome

The repository should ultimately demonstrate:

- production-style Go control-plane engineering;
- Kubernetes controller/operator development;
- SRE and distributed-systems competence;
- deterministic risk and safety policy;
- observability/evidence provenance;
- failure-domain and blast-radius modeling;
- fault-propagation reasoning and containment;
- LLM/RAG engineering with explicit trust boundaries;
- AI-assisted reliability decisions where abstention/escalation are first-class outcomes;
- ML for detection/ranking rather than authority;
- chaos, rollback, watchdog, and human-override semantics.

That portfolio should support progression from **Cloud/Operations → Platform/Infrastructure Software Engineer → AI Infrastructure / Trustworthy Autonomous Infrastructure R&D**.
