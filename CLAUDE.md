# systemd-config

A Go library that (de)serializes systemd config/unit files. Inspired by
`go-systemd/unit`, but supports duplicate sections (e.g. `Address` in
`.network` files). Pure library — no CLI binary, no container image.

## Code quality & checks (Go)

Every check below must pass before commit and in CI. Wire them into a single
`make check` target so local and CI run the exact same commands. Listed
cheapest-to-run first.

### Formatting
- `gofmt -l .` — must print nothing (fail if it lists any file).
- `goimports -l .` (`golang.org/x/tools/cmd/goimports`) — formatting + import grouping.

### Vet / correctness
- `go vet ./...` — catches printf mismatches, unreachable code, bad struct tags, etc.
- Bundles the `loopclosure` analyzer — flags loop variables captured by goroutines/closures.

### Loop safety
- The `go.mod` directive is `go 1.22` (kept low on purpose: a library's go
  directive is a minimum requirement for consumers). Loop variables are
  per-iteration from 1.22 on, so the classic capture footgun is gone; still
  lint for the remaining pitfalls.
- `golangci-lint` linter `copyloopvar` — flags now-redundant loop-variable copies.
- `go test -race ./...` — race detector; surfaces loop/goroutine data races at test time.

### Static analysis / lint (code quality)
- `golangci-lint run` — aggregated meta-linter. Baseline enabled set:
  `staticcheck`, `govet`, `errcheck`, `ineffassign`, `unused`, `gosimple`, `revive`,
  `misspell`, `bodyclose`, `nilerr`, `gocritic`.
- Error correctness: `errorlint` (proper `%w` wrapping), `wrapcheck` (wrap errors
  at package boundaries).
- Complexity & duplication: `gocyclo`/`cyclop` (cap function complexity), `dupl` (copy-paste).
- Optional but recommended: `gofumpt` (stricter gofmt) and `nilaway` (Uber's nil-panic
  static analysis — experimental, run as a separate step).
- Pin the `golangci-lint` version and commit `.golangci.yml` so results are reproducible.

### Security (SAST + vulnerabilities)
- `gosec ./...` — SAST for insecure patterns.
- `govulncheck ./...` — official Go vulnerability scanner. Reachability-aware: only reports
  CVEs in deps/std-lib that the code actually calls, so few false positives. Run in CI.
- `gitleaks detect --no-git` (and as a pre-commit hook) — block tokens or secrets
  from ever being committed.

### Dependency & supply chain
- `go mod tidy` then fail CI if it dirties the tree: `git diff --exit-code go.mod go.sum`.
- `go mod verify` — module contents match recorded checksums.
- `go-licenses check ./...` — dependency licenses stay compatible with Apache-2.0.
- Automate updates with Dependabot/Renovate, and run `govulncheck` on a daily CI schedule
  (not just on PRs) to catch newly disclosed CVEs.

### Build & tests
- `go build ./...` must stay clean (library — no release binaries to produce).
- Tests and coverage are mandatory and have their own policy — see **Testing** below.

### CI
- One `make check` target running all of the above; CI invokes the same target.
  Install tool versions via `go install ...@<pinned>` or a `tools.go`.
  (CI is currently Travis running `go test -race` + codecov upload; migrate it to
  call `make check` so local and CI never drift.)
- Local hooks run the fast subset before code leaves the machine (e.g. `lefthook.yml`,
  installed via `make hooks`): pre-commit runs `gofmt` + `gitleaks --staged`; pre-push
  runs `go vet`, `golangci-lint`, and `go test -race`. The full suite still runs in CI.

## Testing

Every package ships with tests and the suite must stay green. The project maintains
**all kinds of tests** — not just unit tests — and **always produces a coverage report**.
New features and bugfixes land with the tests that cover them; a fix for a bug includes a
test that fails before the fix and passes after.

### Test kinds (all required, kept current)
- **Unit** — table-driven, fast, hermetic; the default for every package. No real network/disk.
- **Round-trip / integration** — deserialize real-world systemd unit files
  (`.service`, `.network`, `.mount`, …) from a fixtures directory and assert
  serialize(deserialize(x)) behaves as specified, including duplicate sections
  and MS-DOS line endings.
- **Fuzz** — native `go test -fuzz` on the deserializer/lexer (prime
  crash/malformed-input surface); keep a seed corpus in the repo and run fuzzing
  in CI on a schedule.
- **Golden** — snapshot serializer output so format changes are explicit.
- **Benchmarks** — `go test -bench=. -benchmem` for the lexer/parser hot paths;
  watch for regressions.
- **Race** — run the suite with `-race`.

### Coverage report (always produced)
- Generate a profile every run:
  `go test -race -covermode=atomic -coverprofile=coverage.out ./...`.
- Human-readable views: `go tool cover -func=coverage.out` (per-func summary) and
  `go tool cover -html=coverage.out -o coverage.html` (annotated source).
- CI uploads the profile (codecov is already wired up) and **fails below the
  line-coverage minimum in the Makefile** (enforced by `make coverage` via
  `COVERAGE_MIN`, currently 90%; ratchet upward over time). Never let coverage
  silently drop.

## Releases

- Tagging a release means: rename the CHANGELOG `Unreleased` section to the
  version + date, commit, tag `vX.Y.Z`, push the tag, then `gh release create`
  with the changelog section as notes.
- `SECURITY.md` is deliberately version-agnostic ("latest release only") —
  never write concrete version numbers into it; there is nothing to update
  there when releasing.

## License

Apache License 2.0.
