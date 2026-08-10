# Contributing

This project optimizes for safety, explainability and reproducibility rather than maximum autonomy.

For changes that expand actuation capability, include:

- threat-model impact;
- authority/risk classification;
- parameter validation;
- preconditions and verification criteria;
- rollback semantics;
- tests for bypass/fail-closed behavior;
- audit/provenance fields;
- an ADR when a trust boundary changes.

Do not add generic model-generated shell execution as a shortcut.
