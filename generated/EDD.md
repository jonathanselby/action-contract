# Action Contract — Implementation-Derived Engineering Evidence

**Source:** published Go source, tests and CI workflow
**Evidence commit:** `0cab516b28ec724936b7acedb4651b2cf5637333`
**Target:** v0.1.0, Go 1.26
**Evidence status:** source fidelity verified against local blob hashes; remote Go 1.26 CI completion must be checked independently

## Recovered component boundary

```text
HUMAN / API / MCP / EVENT (lightweight example translation)
                      |
               trusted Principal
                      |
                      v
          Action[typed input] + Meta
                      |
                      v
          actioncontract.Run(ctx, Definition, Action, execute)
                | validate trusted definition
                | exact type/version match
                | require governed references when configured
                | forward original ctx and action unchanged
                v
            application function
                |
                v
       Result[typed output] + explicit Outcome
       + optional evidence/event references
                |
          validate consistency, replace result metadata
                |
                v
          initiator-specific response
```

The public root package is `contract.go` and `version.go`; no internal transport, persistence, dispatch, registry, security provider, global state or worker exists. The single application domain example lives in `internal/exampleapp/app.go`. Four translation-only functions in `internal/exampleinitiators/adapters.go` share one `run` helper and the same `exampleapp.Execute` function. `examples/initiators/main.go` invokes all four without any network or broker.

## Recovered public API

- `Meta`: type, version, actor, authority, correlation and causation references.
- `Action[I]`: generic input plus `Meta`.
- `Result[O]`: generic output, `Meta`, `Outcome`, optional evidence and event references.
- `Definition`: trusted expected action type/version and `Governed` flag.
- `Outcome`: `success`, `refused`, `failed`, `indeterminate`.
- `Run[I,O](context.Context, Definition, Action[I], func(context.Context, Action[I]) (Result[O], error))`: the only invocation helper.
- `Version`: package version `0.1.0`.

The runtime core imports only `context`, `errors` and `strings` from the standard library. There are no third-party modules in `go.mod`.

## Recovered validation and execution

1. Reject invalid local definitions, nil context or nil executor with `failed` and a typed error; no application execution.
2. Reject absent or mismatched action type/version with `refused` and no execution.
3. For a `Governed` definition, reject blank actor or authority references with `refused` and no execution.
4. Call the supplied application function once with the **original** context and action, including input metadata and typed input.
5. Validate the returned outcome. An unrecognised outcome or a `success`/`refused` with a non-nil error becomes `indeterminate`, with `ErrInvalidOutcome` and no claimed output/evidence/events. Diagnostics are retained with `errors.Join`.
6. For valid outcomes, preserve the executor's output and optional evidence/event references and overwrite result `Meta` with the validated incoming action `Meta`.

`Run` does not interpret an error as a fifth outcome, infer external commit state, retry work or map `context.Canceled` to an assumed business failure. An executor that cannot determine the outcome of a remote side effect must return `indeterminate`.

## Recovered authority and state boundaries

`Definition` is application-selected, never an initiator-controlled registration. Governed actor/authority validation is only a presence check. The host must resolve identity/authority through trusted authentication/authorisation and must enforce actual domain rights on protected actions. Example adapter structs contain deliberately untrusted claims which are ignored; all authoritative identifiers are taken only from the host-supplied `Principal` argument.

No authentication/IAM, permission store, token parsing, routing, persistence, retry, queue, background goroutine, model runtime, event publication, tracing platform or mutable package state exists. Correlation and causation remain caller-managed opaque strings. Events and evidence are references only.

## Recovered executable evidence

| Locked PRD acceptance | Published test / source evidence |
|---|---|
| 1. Three initiators, same domain behaviour | `TestMultipleInitiatorsSameExecution` exercises human/API/MCP/event through one domain function |
| 2. Typed type/version validation | `TestActionMetadataValidation` including incorrect/blank version and trusted-definition configuration |
| 3. Actor, authority, correlation, causation preserved | `TestContextMetadataPreserved`; `TestMultipleInitiatorsSameExecution` |
| 4. Distinct outcomes | `TestOutcomeClassification` including contradiction and unknown outcome |
| 5. Evidence references | `TestEvidenceReferences` |
| 6. Event references | `TestEventReferences` |
| 7. Context cancellation/deadline | `TestContextCancellation`; `TestExampleDomainRefusalAndCancellation` |
| 8. Invalid metadata refused before execution | `TestInvalidMetadataRefused` |
| 9. No transport/provider dependency | `TestCoreDependencyBoundary`, `go.mod`, import review |
| 10. Vet, race, native, Linux ARM64 | `.github/workflows/test.yml` declares Go 1.26 formatting, vet, race, native and ARM64 gates |

Additional evidence: `TestForgedAuthorityCannotElevate`, `TestInvalidConfigurationFailsClosed`, `TestConcurrentExecution`, `TestVersionMatchesFile`, and non-gating `BenchmarkDirectExecute`/`BenchmarkRun`. The example has no transport framework or external infrastructure.

The local verification used an identical source tree except for temporarily changing the **test copy's** `go.mod` from Go 1.26 to Go 1.23 for the installed compiler. It passed `go mod tidy`, `go vet ./...`, `go test -race -count=1 ./...`, five repeat test runs, native build, Linux ARM64 cross-build and four-initiator example execution. Under Go 1.23.2, the direct benchmark was about 10–12 ns/op and `Run` about 61–62 ns/op, both 0 B/op and 0 allocs/op. These numbers are not a Go 1.26 performance claim or production latency guarantee.

## Published source fidelity

The repository tree at the evidence commit contains 21 published files. Local Git blob hashes for all 21 files were compared against GitHub's reported tree/blob SHAs and matched exactly. `PRD.md` was not edited. The Go module remains set to `go 1.26`; the alternate Go 1.23 local-check copy was never published.

## Explicit limits

- Example adapters simulate translation only. They are not real HTTP/MCP/Kafka transports and do not prove integration with any deployed product.
- Actor/authority references are neither authenticated nor authorised by `Run`.
- `Run` cannot prevent an application from bypassing it and calling a domain function directly.
- Invalid local contract configuration is classified as `failed` before domain invocation; invalid executor outcomes are conservatively `indeterminate` after possible effects.
- ARM64 cross-compilation is not ARM64 runtime or constrained-hardware benchmark proof.
- No performance or allocation SLA is provided; local benchmarks are informational.
- Remote GitHub Actions Go 1.26 success is a separate release gate and must be observed rather than inferred from the workflow file.
