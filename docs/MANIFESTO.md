# Manifesto

## Premise

Production infrastructure is deterministic enough to demand crisp authorization, while modern ML and LLM systems are probabilistic enough that uncertainty can never be wished away. The safe composition is therefore **probabilistic reasoning behind deterministic authority**.

## Principles

1. **Diagnostic Authority ≠ Remediation Authority.** The ability to understand a system never implies the ability to mutate it.
2. **Reason probabilistically. Authorize deterministically.** Confidence is evidence for a decision, not a permission primitive.
3. **Proposals are data, not code.** Models emit typed intent. Trusted executors own implementation.
4. **Human override has absolute precedence.** Once an operator intervenes, automation relinquishes ownership immediately.
5. **The brake is outside the accelerator's failure domain.** Watchdogs and rollback paths must not depend on the planner remaining healthy.
6. **No autonomous policy drift.** Learned correlations, pseudo-labels, and remediation outcomes may improve ranking; they may not silently expand permissions.
7. **Minimum perturbation first.** Prefer passive evidence, simulation, dry-run, canary, and the smallest reversible change.
8. **Verification is independent.** The component proposing a repair does not get to declare its own success.
9. **Provenance is part of correctness.** Every diagnosis and action must be explainable from evidence, policy, identity, model/version, approval, and observed outcome.
10. **Fail closed.** Unknown action, malformed proposal, stale evidence, missing rollback, policy ambiguity, or unavailable watchdog means no mutation.

## What success looks like

A trustworthy self-healing system should be able to answer, after every incident:

- What happened?
- What evidence supported the diagnosis?
- What contradictory evidence existed?
- Why was this action considered?
- Which deterministic policy rule allowed or denied it?
- What was the expected blast radius?
- Who or what approved it?
- What exactly changed?
- How was success verified independently?
- What rollback path existed and was it exercised?
- What, if anything, was learned without changing safety authority?
