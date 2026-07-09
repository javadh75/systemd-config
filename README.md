# systemd-config

[![Go Reference](https://pkg.go.dev/badge/github.com/javadh75/systemd-config.svg)](https://pkg.go.dev/github.com/javadh75/systemd-config)
[![CI](https://github.com/javadh75/systemd-config/actions/workflows/ci.yml/badge.svg)](https://github.com/javadh75/systemd-config/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/javadh75/systemd-config/branch/master/graph/badge.svg?token=OJLajmXJv4)](https://codecov.io/gh/javadh75/systemd-config)
[![Go Report Card](https://goreportcard.com/badge/github.com/javadh75/systemd-config)](https://goreportcard.com/report/github.com/javadh75/systemd-config)

A simple systemd config (de)serializer.

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

	// Look up a value (last assignment wins, like systemd).
	if network := unit.SectionByName("Network"); network != nil {
		if dns, ok := network.Value("DNS"); ok {
			fmt.Println("DNS:", dns)
		}
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

## Behavior notes

- **Duplicate sections and options** are preserved in order. `Section.Value`
  follows systemd's last-assignment-wins rule; `Section.Values` returns every
  occurrence.
- **Comments and blank lines are not preserved**: deserializing discards
  them, so a deserialize/serialize round trip produces a normalized file.
- **Continuation lines** follow systemd.syntax(7): a line ending in `\` is
  joined with the following non-comment line and the backslash becomes a
  space. A value therefore cannot end in a backslash — dangling markers are
  dropped.
- **Empty sections survive** a round trip (`[Install]` with no options stays).
- **Line length** is capped at 1 MiB per line (systemd's `LONG_LINE_MAX`);
  longer lines yield `ErrLineTooLong`.
- **Canonical output**: serializing a parsed unit yields a fixpoint —
  parsing and serializing the output again reproduces it byte for byte
  (fuzz-tested).

## Development

```sh
make check     # full quality gate: tidy, fmt, vet, lint, security, tests
make coverage  # coverage report, fails below 80%
make hooks     # install pre-commit/pre-push git hooks (lefthook)
make fuzz      # short fuzz run of the deserializer
```

See `CLAUDE.md` for the full tooling and testing policy.

## License

Apache License 2.0.
