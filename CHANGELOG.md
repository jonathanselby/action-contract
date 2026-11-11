# Changelog

## 0.1.0 — Initial release candidate

- Typed `Action[I]` and `Result[O]` with explicit identity/version, actor/authority and correlation/causation references.
- Four outcomes: success, refused, failed and indeterminate.
- One function-based `Run` boundary, with validated application-owned definition and governed-reference presence checks.
- Conservative result consistency, original-context propagation and caller metadata preservation.
- Evidence and domain-event references without persistence or event transport.
- Human/API/MCP/event demonstration adapters sharing one application executor.
- Adversarial acceptance, race, standard-library dependency, build and ARM64 CI gates.

### Design lineage (EDD v0.1 → v0.2)

- Replaced the mandatory executor interface with one generic typed function and trusted action definition.
- Defined full typed action/result envelopes and fail-conservative outcome handling.
- Made authority provenance, metadata retention, cancellation and negative security tests explicit.
- Consolidated the hand-written EDD without changing locked PRD v0.1 requirements.
