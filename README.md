# systemd-config

[![Go Reference](https://pkg.go.dev/badge/github.com/javadh75/systemd-config.svg)](https://pkg.go.dev/github.com/javadh75/systemd-config)
[![CI](https://github.com/javadh75/systemd-config/actions/workflows/ci.yml/badge.svg)](https://github.com/javadh75/systemd-config/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/javadh75/systemd-config/branch/master/graph/badge.svg)](https://codecov.io/gh/javadh75/systemd-config)

A simple systemd config (de)serializer. It parses and writes unit/config
files, and can compute the effective configuration of a unit combined
with its drop-ins (see [Drop-ins](#drop-ins)).

This project is highly inspired by
[go-systemd/unit](https://github.com/coreos/go-systemd/tree/main/unit). The
difference is duplicate-section support: unlike go-systemd/unit,
systemd-config handles repeated sections such as the `[Address]` sections in
`.network` files.

## Install

```sh
go get github.com/javadh75/systemd-config
```

## Usage

```go
package main

import (
	"fmt"
	"log"
	"os"

	systemdconfig "github.com/javadh75/systemd-config"
)

func main() {
	f, err := os.Open("eth0.network")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	unit, err := systemdconfig.Deserialize(f)
	if err != nil {
		log.Fatal(err)
	}

	// Look up a value (last assignment wins, like systemd, even across
	// duplicate sections).
	if dns, ok := unit.Value("Network", "DNS"); ok {
		fmt.Println("DNS:", dns)
	}

	// Duplicate sections are preserved and addressable.
	for _, addr := range unit.SectionsByName("Address") {
		if v, ok := addr.Value("Address"); ok {
			fmt.Println("address:", v)
		}
	}

	// Edit and write back.
	unit.AddSection("Route").AddOption("Gateway", "10.0.0.1")
	if _, err := unit.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
```

## Drop-ins

`Merge` computes the effective configuration of a unit combined with its
drop-ins (`systemctl cat` semantics): drop-ins apply in order as if
appended, and an empty assignment (`ExecStart=`) resets every earlier
occurrence of that option, as described in systemd.unit(5). Duplicate
sections are never collapsed.

```go
base, _ := systemdconfig.Deserialize(baseFile)     // nginx.service
override, _ := systemdconfig.Deserialize(dropFile) // nginx.service.d/override.conf

effective := systemdconfig.Merge(base, override)
cmd, _ := effective.Value("Service", "ExecStart")
```

## Behavior notes

- **Duplicate sections and options** are preserved in order. `Unit.Value`
  follows systemd's last-assignment-wins rule across duplicate sections;
  `Unit.Values` returns every occurrence. `Section.Value`/`Section.Values`
  do the same within a single section.
- **Comments and blank lines are not preserved**: deserializing discards
  them, so a deserialize/serialize round trip produces a normalized file.
- **Continuation lines** follow systemd.syntax(7): a line ending in `\` is
  joined with the following non-comment line and the backslash becomes a
  space. A value therefore cannot end in a backslash — dangling markers are
  dropped.
- **Empty sections survive** a round trip (`[Install]` with no options stays).
- **Line length** is capped at 1 MiB per line (systemd's `LONG_LINE_MAX`);
  longer lines yield `ErrLineTooLong`.
- **Assignments before the first section header are rejected** with
  `ErrAssignmentOutsideSection`, as in systemd.
- **Canonical output**: serializing a parsed unit yields a fixpoint —
  parsing and serializing the output again reproduces it byte for byte
  (fuzz-tested).

## Development

```sh
make check     # full quality gate: tidy, fmt, vet, lint, security, tests
make coverage  # coverage report, fails below COVERAGE_MIN (see Makefile)
make hooks     # install pre-commit/pre-push git hooks (lefthook)
make fuzz      # short fuzz run of the deserializer
```

See `CLAUDE.md` for the full tooling and testing policy.

## License

Apache License 2.0.
