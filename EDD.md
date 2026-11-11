# Action Contract — Engineering Design Document

**Version:** 0.2  
**Governing PRD:** `PRD.md`, version 0.1 (locked)  
**Status:** Approved implementation baseline; test and build results are recorded separately in CI and generated evidence.

## 1. Purpose and boundaries

> **Different initiators. One application action contract.**

Action Contract is a small, in-process Go execution boundary. Human UI, API, MCP, event, model and browser-automation initiators translate their inputs into the **same typed application action**; application behaviour is invoked once through the same contract and its result is translated back by the initiator.

The component owns **action/result types, deterministic shared-metadata validation, result consistency, and context propagation**. It does not own authentication, authorisation policy, domain validation, transport, routing, orchestration, retries, persistence, scheduling, event publication, model execution or browser automation.

**Engineering constraints:** standard-library-only core; no registry, dispatcher, middleware, reflection, queue, background worker or hidden state; one comprehensible execution path. Do not add capacity or latency limits without evidence that the in-process contract needs them.

## 2. Architecture and execution flow

```text
 HUMAN UI   API   MCP   EVENT   MODEL   BROWSER
      \      |     |      |       |      /
       +-----+-----+------+-------+-----+
                        |
                        v
           APPLICATION SECURITY LAYER
        (resolves trusted actor/authority)
                        |
                        v
                 TYPED ACTION
                        |
                        v
           ACTION CONTRACT: Run(...)
             | validate identity
             | validate required references
             | forward context unchanged
             v
             APPLICATION FUNCTION
             | domain rules and effects
             v
             TYPED RESULT + OUTCOME
             | validate outcome consistency
             | retain action metadata
             v
               INITIATOR RESPONSE
```

Initiator adapters **only translate** requests and responses. They do not define alternative domain execution paths or independently grant authority. The application integrates its own security checks before invoking the contract and retains responsibility for domain-level permission checks.

## 3. Public Go contract

The intended public contract consists of the following small set of types and **one helper**. A mandatory `Executor` interface is unnecessary: ordinary Go functions and method values are sufficient.

```go
package actioncontract

import "context"

type Meta struct {
    Type          string
    Version       string
    ActorID       string
    AuthorityID   string
    CorrelationID string
    CausationID   string
}

type Action[I any] struct {
    Meta  Meta
    Input I
}

type Outcome string

const (
    Success       Outcome = "success"
    Refused       Outcome = "refused"
    Failed        Outcome = "failed"
    Indeterminate Outcome = "indeterminate"
)

type EvidenceRef struct {
    ID   string
    Kind string
}

type EventRef struct {
    ID   string
    Type string
}

type Result[O any] struct {
    Meta     Meta
    Outcome  Outcome
    Output   O
    Evidence []EvidenceRef
    Events   []EventRef
}

// Definition is supplied by trusted application code, not the initiator.
// Governed requires both ActorID and AuthorityID to be present.
type Definition struct {
    Type     string
    Version  string
    Governed bool
}

func Run[I, O any](
    ctx context.Context,
    definition Definition,
    action Action[I],
    execute func(context.Context, Action[I]) (Result[O], error),
) (Result[O], error)
```

Concrete applications define `I` and `O` using ordinary Go types. `Definition` is an application-owned value, **not** a registration system, schema registry or authorisation policy. The function may be a domain method value; no adapter interface is required in the core.

## 4. Validation, authority and outcomes

### Validation before execution

`Run` performs the following deterministic checks in order:

1. Reject a missing/invalid **application-owned** definition, nil context or nil executor as a **known contract/configuration failure**: `failed` plus non-nil error; executor not called.
2. Compare action `Meta.Type` and `Meta.Version` exactly with the trusted definition. Missing, blank or mismatched values produce `refused`, nil error; executor not called. The initiator cannot choose the expected identity/version.
3. If `Definition.Governed` is true, require non-blank `ActorID` **and** `AuthorityID`. Missing references produce `refused`, nil error; executor not called. For other actions, these references are optional.
4. Pass the original context and typed action to `execute`. Domain input validation, identity verification and permission enforcement are **not** implemented by this component.

Reference-presence checks are not authentication or authorisation. In particular, a populated `AuthorityID` **does not prove authority**. The application must obtain trusted actor and authority values from its security layer rather than copying user-, agent- or MCP-supplied identity claims. Applications must not expose an alternative ungoverned path to the same protected operation.

### Explicit outcome semantics

| Outcome | Meaning |
|---|---|
| `success` | Application confirms successful completion. |
| `refused` | A deterministic validation or business rule declined the action. |
| `failed` | Application confirms execution failed; it must not claim failure if an external effect might have completed. |
| `indeterminate` | Application cannot establish the final external effect or outcome. |

An `error` conveys operational/contract diagnostic information, **not** a fifth outcome. A deterministic refusal normally returns no error. A valid `failed` or `indeterminate` outcome may carry an execution error. A successful or refused outcome paired with a non-nil execution error is contradictory. An absent/unrecognised outcome is also invalid.

After calling `execute`, `Run` must:

- accept only the four defined outcome values;
- return a contradictory/invalid executor result **conservatively as `indeterminate` with a non-nil contract error**, never silently as success or definitive failure;
- preserve a valid `indeterminate` classification even when the accompanying error is a timeout or cancellation;
- set `Result.Meta` from the **validated incoming action**, not from mutable metadata supplied by the executor; and
- retain explicitly returned evidence/event references for valid results. An invalid result must not be presented as a successful business output; diagnostics may be used to reconcile external effects.

When an executor reports `failed`, it is responsible for knowing the effect did not succeed. When that cannot be established, it must report `indeterminate`. `Run` cannot discover a remote transaction's state and must not retry it.

### Cancellation and deadlines

`Run` forwards the supplied `context.Context` **unchanged**, including its cancellation and deadline, to the executor; it does not replace it, start a goroutine or intercept a cancellation to infer the transaction's final state. A pre-cancelled context reaches the executor. If cancellation occurs during an external effect, the executor must report `indeterminate` unless completion or non-completion is known.

## 5. Metadata, evidence and events

`CorrelationID` groups related work; `CausationID` identifies the immediately preceding action or event where known. Both are optional opaque references; the component neither generates IDs nor performs distributed tracing.

The validated `Meta` reaches the executor unchanged and is copied to the returned `Result`, retaining type/version, actor, authority, correlation and causation across invocation. The executor cannot rewrite those result references through the contract.

`EvidenceRef` identifies evidence created or used by the application without storing it. `EventRef` identifies domain events actually produced by application behaviour without publishing or guaranteeing delivery. References do not imply an evidence store, event broker, outbox or transaction manager. Applications must not claim an uncertain event was emitted merely because a transaction was attempted.

## 6. Failure handling and deliberate exclusions

| Condition | Contract behaviour |
|---|---|
| Invalid application definition, nil context/executor | `failed` and error; no execution |
| Missing/mismatched action type or version | `refused`; no execution |
| Missing required actor/authority reference | `refused`; no execution |
| Malformed initiator-specific input | Rejected by adapter before `Run` |
| Deterministic application/domain refusal | `refused` result |
| Confirmed operational execution failure | `failed` result, with error when available |
| Unconfirmed external effect, including ambiguous timeout | `indeterminate` result |
| Invalid/contradictory executor outcome | `indeterminate` and contract error |
| Caller cancellation/deadline | Forwarded to executor; classification belongs to executor |

There is no retry policy, distributed execution, persistence, durable queue, schema registry, cross-process discovery, transport security or application framework. Failures are explicit rather than hidden in middleware or mutable internal state.

## 7. Tests and PRD acceptance evidence

The **locked PRD v0.1** has eight functional requirements and ten acceptance criteria. The implementation must map each criterion to named executable evidence, not merely documentation.

| PRD acceptance criterion | Required executable evidence |
|---|---|
| 1. Same behaviour from 3+ initiators | `TestMultipleInitiatorsSameExecution`: human/API/MCP/event examples use the same typed action and application function |
| 2. Stable, validated type/version | `TestActionMetadataValidation`: valid, absent, mismatched and untrusted expected-identity cases |
| 3. Actor, authority, correlation, causation survive | `TestContextMetadataPreserved`: executor receives references; result reflects validated incoming metadata |
| 4. Four distinguishable outcomes | `TestOutcomeClassification`: four outcomes, invalid enum, contradictory error/result and ambiguous timeout |
| 5. Evidence references | `TestEvidenceReferences`: references returned without evidence implementation |
| 6. Domain-event references | `TestEventReferences`: references returned without broker or publishing |
| 7. Context cancellation reaches executor | `TestContextCancellation`: original context, pre-cancellation and deadlines; indeterminate after uncertain effects |
| 8. Invalid metadata deterministically refused | `TestInvalidMetadataRefused`: never invokes application function, including governed-action missing references |
| 9. No transport/provider coupling | `TestCoreDependencyBoundary`: core import/dependency check; only Go standard library |
| 10. Vet, race, native and ARM64 build | `.github/workflows/test.yml`: Go 1.26, formatting, `go vet ./...`, `go test -race ./...`, native `go build ./...`, `GOOS=linux GOARCH=arm64 go build ./...` |

Additional negative/security cases within those tests: forged actor/authority claims rejected by the example application's security boundary; invalid trusted definition; nil executor/context; attempted executor metadata rewrite; error with `success`/`refused`; malformed and unknown outcomes; failure vs indeterminate under uncertainty. Test adapters without implementing real HTTP, MCP or event transports.

Include a small no-allocation/latency benchmark comparing direct function invocation with the helper. Record measurements under Go 1.26 in the target build environment; **do not assert a fixed nanosecond guarantee**. The core has no network, I/O, locks, background processing, reflection or deliberate per-call allocation. Race testing remains mandatory even if the helper has no shared mutable state.

## 8. Build, publication and definition of done

- One small Go module/package, with standard-library-only core; use the existing Garage component repository conventions.
- Build and test **offline before publishing**. Publish in a few coherent commits, not recursive micro-updates.
- CI must prove formatting, vet, tests including race, native build, and Linux ARM64 cross-build using Go 1.26. ARM64 *build proof* does not imply runtime testing on ARM64.
- Verify uploaded GitHub source against the tested local source and confirm CI results.
- After the final source-changing commit, generate the **code-derived EDD and comparison** to this approved design; do not use a prospective document as evidence of implementation.
- GitHub About, topics and manual release publication remain manual. Provide paste-ready text upon completion.

**Definition of done:** all ten PRD acceptance criteria pass with named evidence; the common execution path is understandable in one sitting; no initiator-specific domain semantics, authority escalation, incorrect certainty after external timeouts, transport coupling, or unjustified framework machinery has been introduced.
