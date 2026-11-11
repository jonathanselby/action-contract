# Action Contract — Release Checklist

## Contract

- [ ] Locked `PRD.md` v0.1 retained byte-for-byte.
- [ ] Approved `EDD.md` v0.2 aligned with implementation; all eight functional requirements and ten acceptance criteria covered.
- [ ] `VERSION`, `actioncontract.Version`, changelog and eventual release tag agree.
- [ ] Four initiator examples converge on the same typed application function.
- [ ] Trusted definition, governed-reference presence and action metadata checks are enforced before execution.
- [ ] No authentication, authorisation policy, transport, registry, workflow, retry, persistence or event publication platform is embedded.
- [ ] Invalid and contradictory executor outcomes cannot become successful or definitively failed external actions.
- [ ] Evidence/events remain application-owned references; original context and metadata survive.

## Engineering gate

- [ ] `go mod tidy` produces no diff; no third-party module dependencies.
- [ ] `gofmt` clean; `go vet ./...` passes.
- [ ] `go test -race -count=1 ./...` passes.
- [ ] Go 1.26 native build, example execution and Linux ARM64 cross-build pass.
- [ ] TestCoreDependencyBoundary and forged-authority example tests pass.
- [ ] Direct-call versus Run benchmark is recorded, without inventing an SLA.

## Evidence and release

- [ ] Published GitHub files match the tested offline source.
- [ ] GitHub Actions `test` green at final source-changing commit.
- [ ] `generated/EDD.md` pinned to that source commit.
- [ ] `generated/COMPARISON.md` records any differences; no unexplained mismatch.
- [ ] Documentation-only evidence commit has green CI.
- [ ] Only then, manually create the immutable `v0.1.0` tag and GitHub release.
