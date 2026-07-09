# Changelog

## Unreleased

### Breaking changes

- `Deserialize` now rejects assignments (or any non-comment text) that
  appear before the first section header with
  `ErrAssignmentOutsideSection`, matching systemd. Previously such text
  was silently discarded.

### Added

- `Unit.Value(section, option)` — one-call lookup that treats duplicate
  sections as one merged section with last-assignment-wins semantics,
  as systemd does.
- `Unit.String()` — the canonical serialized form, implementing
  `fmt.Stringer`.
- Testable examples (`ExampleDeserialize`, `ExampleUnit_Value`,
  `ExampleUnit_WriteTo`) rendered on pkg.go.dev.
- Real-world fixtures under `testdata/` (services, sockets, netdev,
  network, mount, timer, daemon configs, drop-ins, CRLF) with golden
  round-trips and expected-value tests.
- Fuzzing in CI: a 15-second smoke run on every push/PR and a nightly
  10-minute run; the fuzzer is now also seeded with every testdata
  fixture.

### Changed

- The lexer is synchronous: the goroutine/channel pipeline inherited from
  go-systemd/unit is gone. This removes a latent goroutine leak on
  early-error returns and makes `Deserialize` ~2.2x faster
  (63.6µs → 29.3µs, 2000 → 1804 allocs/op on the parser benchmark).
- `go.uber.org/goleak` (test-only dependency) now fails the suite if any
  test leaks a goroutine.
- README: removed the Go Report Card badge (the service is retired) and
  updated the usage example to `Unit.Value`. CI now authenticates codecov
  uploads (`CODECOV_TOKEN`), unsticking the stale coverage badge.

## v0.3.0 — 2026-07-09

### Breaking changes

- `SYSTEMD_LINE_MAX` renamed to `LineMax` and raised from 2048 to 1 MiB,
  matching modern systemd's `LONG_LINE_MAX`.
- `SYSTEMD_NEWLINE` removed (unused).
- `NewLexer` unexported: it returned unexported types and was unusable
  outside the package. Use `Deserialize`.
- `NewUnitOption` renamed to `NewOptionValue` to match the type it builds.
- `InitialCompareSliceGenerator` and `AllAreTrue` unexported (internal
  comparison helpers).
- `WriteNewLine`, `WriteSectionHeader`, `WriteOptionValue` unexported
  (serialization internals). Use `Serialize` or `Unit.WriteTo`.
- Continuation lines now follow systemd.syntax(7): a trailing `\` is
  replaced by a space and joined with the next non-comment line
  (previously the backslash and a newline were kept in the value).
  Dangling continuation markers at end of value/file are dropped.
- `Serialize` no longer drops sections that have no options.
- go directive is now 1.22 (minimum supported Go).

### Added

- `Unit.SectionsByName`, `Unit.SectionByName`, `Unit.AddSection`.
- `Section.AddOption`, `Section.Value` (last-assignment-wins),
  `Section.Values`.
- `Unit.WriteTo` implementing `io.WriterTo`.
- Comment lines interleaved inside continuation lines are skipped, as in
  systemd.
- Quality tooling: Makefile gate (`make check`), golangci-lint v2 config,
  lefthook hooks, GitHub Actions CI (replacing Travis), scheduled
  govulncheck, Dependabot.
- Round-trip golden tests, fuzzing with seed corpus, and benchmarks;
  coverage gate at 80% (currently ~96%).

### Fixed

- `Section.Match`/`Unit.Match` failed on collections containing identical
  duplicate entries: matched elements are now consumed so duplicates
  compare correctly.
- Serialized output is canonical: parsing and re-serializing it is a
  byte-for-byte fixpoint (fuzz-verified).
