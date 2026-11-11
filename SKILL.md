# Action Contract — Skill

## Use this when

An application needs the same typed, governed action semantics across human, API, MCP, event, model or browser initiators.

## Core rule

> **Different initiators. One application action contract.**

## Change discipline

1. Preserve the locked `PRD.md` v0.1 and approved `EDD.md` v0.2; all eight functional requirements and ten acceptance criteria must remain covered.
2. Keep a single typed `Run` helper. Do not add an action registry, command bus, dispatcher, middleware chain or workflow runtime.
3. Initiator-specific adapters only translate; they must converge on the same application-owned executor and action definition.
4. Expected action type/version and `Governed` are selected by trusted application code, never by the initiator.
5. Actor/authority references must be resolved by the host security layer. A populated reference is not proof of permission; application domain checks remain mandatory.
6. Refuse invalid action metadata before executing. Treat nil context/executor or invalid trusted definitions as known contract failures.
7. Preserve the original context and input metadata. Result metadata cannot be rewritten by the executor.
8. Preserve all four outcomes. An unknown external effect remains `indeterminate`; contradictory outcomes become `indeterminate` with a contract error, never success.
9. Keep evidence and event references opaque. Do not implement their stores, publication or delivery guarantees.
10. Standard-library-only runtime; avoid reflection, queues, goroutines and hidden state. Benchmark the helper without declaring a fixed latency SLA.
11. Prove all named acceptance tests, security negatives, format/vet/race, native and Linux ARM64 build under Go 1.26 CI.
12. Build and test offline, publish in few coherent commits, verify GitHub source fidelity, then pin code-derived evidence to the final source-changing commit. No recursive publication experiments.
