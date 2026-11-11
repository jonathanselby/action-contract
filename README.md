# Action Contract

[![test](https://github.com/jonathanselby/action-contract/actions/workflows/test.yml/badge.svg)](https://github.com/jonathanselby/action-contract/actions/workflows/test.yml)

A small, transport-independent Go contract for invoking the same application action from human, API, MCP, event, model or browser initiators.

> **Different initiators. One application action contract.**

The package provides typed action/result envelopes, stable type/version validation, actor/authority references, correlation/causation, four explicit outcomes and optional evidence/event references. It is not a command bus, transport adapter, security engine or workflow framework.

## Run the examples

Requires Go 1.26 for the supported build; the core uses only the Go standard library.

```sh
go run ./examples/initiators
```

The example prints `success` four times: the human, API, MCP and event inputs translate to the same typed action and application function. There are no HTTP servers, MCP transports, event brokers or external services in this example.

## Public Go contract

```go
type Input struct { CaseID string }
type Output struct { Approved bool }

definition := actioncontract.Definition{
    Type: "case.approve", Version: "1", Governed: true,
}
action := actioncontract.Action[Input]{
    Meta: actioncontract.Meta{
        Type: "case.approve", Version: "1",
        ActorID: trustedActorID, AuthorityID: trustedAuthorityID,
        CorrelationID: correlationID,
    },
    Input: Input{CaseID: "123"},
}
result, err := actioncontract.Run(ctx, definition, action,
    func(ctx context.Context, a actioncontract.Action[Input]) (actioncontract.Result[Output], error) {
        // Application owns authorisation, validation and business effects.
        return actioncontract.Result[Output]{
            Outcome: actioncontract.Success,
            Output: Output{Approved: true},
        }, nil
    },
)
```

The application must resolve `trustedActorID` and `trustedAuthorityID` through its own security layer, not from untrusted client, MCP or agent claims. A populated authority reference is **not** authorisation; domain permissions remain the application's responsibility. A `refused` result is not automatically an HTTP error.

## Outcomes and failure ownership

- `success`: confirmed completed action.
- `refused`: deterministic refusal (invalid action metadata or domain rules).
- `failed`: known failure; never assert that a possibly completed external effect failed.
- `indeterminate`: final outcome cannot be established, including uncertain remote timeouts.

`Run` forwards the original `context.Context`. It refuses invalid action type/version and missing required governed references *before* calling the application function. Invalid local definitions or nil context/executor return `failed` with error. Contradictory executor outcomes are returned as `indeterminate` with a contract error and without claimed business artefacts. Successful results always carry validated incoming action metadata.

## Architecture and security boundary

```text
Untrusted input -> host authentication/authorisation -> typed action
             -> Action Contract: Run -> application function
             -> typed outcome + evidence/event references -> adapter response
```

The contract never establishes identity, grants permission, sends messages, publishes events or retries remote operations. The host selects expected action identity/version and whether it is governed. The reference-only examples do not provide production authentication or persistence.

## Engineering evidence

- [`PRD.md`](PRD.md) — locked v0.1 requirements.
- [`EDD.md`](EDD.md) — approved simplified v0.2 design.
- [`architecture.txt`](architecture.txt) — component/trust boundaries.
- [`SKILL.md`](SKILL.md) — change discipline.
- [`RELEASE.md`](RELEASE.md) — acceptance and release checklist.
- [`generated/EDD.md`](generated/EDD.md) and [`generated/COMPARISON.md`](generated/COMPARISON.md) — code-derived evidence pinned after final source commit.

CI checks module tidiness, formatting, vet, race/acceptance tests, an initiator example, native build, Linux ARM64 cross-compilation and a non-gating execution benchmark. ARM64 cross-compilation is not ARM64 runtime proof.

## Non-goals

No HTTP/MCP/Kafka transport, registry, orchestration, retries, persistence, IAM, authorisation policy, event publishing, schema registry, distributed tracing, model execution or browser automation.

## Licence

MIT. See [`LICENSE`](LICENSE).
