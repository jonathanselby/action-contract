# Action Contract — Intended Design versus Recovered Implementation

**Design:** `EDD.md` v0.2  
**Locked requirements:** `PRD.md` v0.1  
**Implementation evidence commit:** `0cab516b28ec724936b7acedb4651b2cf5637333`

| Design area | Hand-written intent | Published source / verification | Assessment |
|---|---|---|---|
| Single execution boundary | One generic `Run` function, no framework | `contract.go` exports one typed helper | MATCH |
| Typed action/result | `Action[I]`, `Result[O]`, `Meta` | Concrete generic types and typed example | MATCH |
| Stable action identity | Trusted definition, exact type/version | Exact match; negative tests | MATCH |
| Governed references | Actor/authority presence, external security | `Definition.Governed`; host-supplied example `Principal` | MATCH |
| Metadata fidelity | No executor authority rewrite | Result `Meta` overwritten from action; test | MATCH |
| Four outcomes | Explicit classification, no false certainty | `Outcome` enum, consistency checks | MATCH |
| Contradictory executor result | Conservative unknown outcome, error | `indeterminate` and `ErrInvalidOutcome`; artefacts cleared | MATCH |
| Cancellation and deadlines | Forward `context.Context` unchanged | Direct invocation; cancellation/deadline tests | MATCH |
| Evidence/events | References only | Slice fields in `Result`, no store/broker | MATCH |
| Initiator independence | Human/API/MCP/event converge on domain | Four adapter functions share `run` and `exampleapp.Execute` | MATCH |
| Security boundary | No trust in spoofed identity claims | Example ignores untrusted claims; negative tests | MATCH (reference example only) |
| Runtime dependency budget | Standard-library-only core | `context`, `errors`, `strings`; no module dependencies | MATCH |
| Hidden processing state | No registry, queue, workers, I/O, retry | No such runtime code | MATCH |
| Latency discipline | Minimal helper; benchmark | Local 0-allocation benchmark; Go 1.26 CI benchmark configured | MATCH (local evidence only) |
| Release/source evidence | Pin generated docs to final source | GitHub tree and 21 local Git blob hashes match evidence commit | MATCH |
| CI quality gate | Go 1.26 vet, race, native + ARM64 | `.github/workflows/test.yml` declares gates; Go 1.23.2 equivalent passed locally | PENDING REMOTE CI VERIFICATION |

## Adversarial review findings

1. The examples intentionally ignore spoofable actor and authority claims from initiator-specific payloads and obtain values from a separate trusted `Principal` parameter; a blank principal is refused before domain execution. This is reference evidence, **not** an authentication implementation.
2. Executors can return arbitrary results. `Run` validates outcomes and prevents a contradictory `success` or `refused` plus error from masquerading as successful. It also removes claimed output/evidence/events from invalid results and retains diagnostic errors.
3. The helper does not convert `context.Canceled`/deadlines to definitive business failure after an external effect. The executor must classify uncertainty; tests demonstrate `indeterminate` survives cancellation.
4. The public contract is stateless and transport-independent. No additional registry, policy evaluator, generic middleware or workflow layer was needed to close the tests.
5. The test strategy contains only lightweight example adapters rather than real transport integrations, consistent with the locked PRD's non-transport boundary.

## Clarifications rather than hidden mismatches

- The EDD's generic single-execution helper does not guarantee that a consuming application never bypasses it; application architecture and security review must enforce one entry point.
- `Definition.Governed` requires actor and authority **presence**, not their authenticity or sufficiency.
- `Run` cannot verify an external side effect; it preserves the executor's explicit `indeterminate` classification.
- Generated evidence describes the actual code and locally executed verification. GitHub Actions success under Go 1.26 is not inferred from a checked-in YAML file; it must be separately verified.
- ARM64 evidence currently represents cross-compilation, not runtime exercise on ARM64 hardware.

## Result

**Source-to-design MATCH, with one outstanding release gate: independently confirm the final GitHub Actions Go 1.26 run.** There are no unexplained code-versus-EDD deviations at the source-changing commit. The repository's lock-step PRD, approved EDD, code-derived EDD and comparison remain separate documents.
