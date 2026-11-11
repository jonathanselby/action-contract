# Action Contract --- Product Requirements Document

**Version:** 0.1

## 1. Problem

Garage applications can be initiated through a human interface, API, MCP
tool, workflow, model, browser automation or inbound event. Without a
common execution contract, each entry path can develop its own request,
authority, result and evidence semantics.

That creates duplicated application logic and risks API, MCP, AI and
event behaviour becoming different applications.

## 2. Product objective

Provide a deliberately small Go contract for expressing a governed
application action and its result.

> **Different initiators. One application action contract.**

The V0.1 product question is: can human, machine and agent-initiated
work share one typed execution boundary without creating a command bus,
workflow engine or orchestration framework?

## 3. Users

Primary users are Garage applications and reusable components that need
to invoke application behaviour consistently from different interfaces.

## 4. In scope

-   typed action identity and version;
-   initiating actor reference;
-   authority context reference;
-   correlation and causation identifiers;
-   typed input and typed result patterns;
-   explicit success, refusal, failure and indeterminate outcomes;
-   evidence references;
-   emitted domain-event references;
-   context cancellation and deadlines;
-   deterministic validation;
-   example adapters for human/API/MCP/event initiation;
-   tests proving initiator-independent application behaviour;
-   Go 1.26 and `linux/arm64` build proof.

## 5. Out of scope

Workflow orchestration, message transport, event brokering, HTTP
routing, MCP transport, authentication, authorisation policy, retries,
distributed transactions, plugin loading, persistence, scheduling, model
routing and browser automation.

The contract carries identity and authority context; it does not decide
either.

## 6. Functional requirements

### FR1 --- Typed action identity

Every action has a stable type and version. Implementations should use
concrete Go types rather than an unbounded generic payload where
practical.

### FR2 --- Actor and authority context

Every governed action can carry the initiating actor and the authority
context supplied by the application security layer.

### FR3 --- Correlation and causation

Actions and results can carry correlation and causation identifiers so
work can be traced across API, MCP, events and asynchronous execution.

### FR4 --- Explicit outcomes

A result distinguishes success, refusal, failure and indeterminate
outcome. Unknown transaction state must not be silently converted to
success or failure.

### FR5 --- Evidence

A result can reference evidence created or relied upon during execution
without embedding an evidence store in this component.

### FR6 --- Events

A result can expose domain events produced by successful application
behaviour without publishing those events itself.

### FR7 --- Cancellation

Execution contracts propagate `context.Context` so caller cancellation
and deadlines reach application behaviour.

### FR8 --- No transport coupling

The contract must compile and operate without HTTP, MCP, Kafka, SQL or
model-provider dependencies.

## 7. Acceptance criteria

V0.1 is accepted when executable evidence demonstrates:

1.  the same application action can be invoked from at least three
    example initiators without changing domain behaviour;
2.  action type/version are stable and validated;
3.  actor, authority, correlation and causation context survive
    invocation;
4.  success, refusal, failure and indeterminate outcomes are
    distinguishable;
5.  evidence references can be returned without coupling to an evidence
    implementation;
6.  domain events can be returned without coupling to an event
    transport;
7.  context cancellation reaches the executor;
8.  invalid action metadata is deterministically refused;
9.  the package has no HTTP, MCP, Kafka, SQL or model-provider
    dependency;
10. `go vet`, race tests, native build and `linux/arm64` build pass.

## 8. Definition of done

V0.1 is complete when the public contract is small enough to understand
in one sitting, examples prove multiple initiators converge on the same
application behaviour, every acceptance criterion has named executable
evidence, and the component has not grown into a command bus or workflow
framework.
