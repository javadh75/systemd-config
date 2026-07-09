COVERPROFILE := coverage.out
COVERAGE_MIN := 80
PKG          := ./...
GO           := go

# Pinned dev-tool versions (keep in sync with CLAUDE.md).
GOLANGCI_VERSION := v2.12.2
GOSEC_VERSION    := latest
GOVULN_VERSION   := latest
GITLEAKS_VERSION := latest

.PHONY: all check build fmt vet lint security gosec vuln secrets test coverage \
        fuzz bench tidy tools hooks clean

## all: default target — run the full gate
all: check

## check: full quality gate (the same command CI runs)
check: tidy fmt vet lint security test

## build: the library must compile cleanly
build:
	$(GO) build $(PKG)

## fmt: fail if any file is not gofmt-clean
fmt:
	@files=$$(gofmt -l .); if [ -n "$$files" ]; then echo "gofmt needed:"; echo "$$files"; exit 1; fi

## vet: go vet
vet:
	$(GO) vet $(PKG)

## lint: golangci-lint (v2)
lint:
	golangci-lint run

## security: SAST + vulnerabilities + secret scan
security: gosec vuln secrets

gosec:
	gosec -quiet ./...

vuln:
	govulncheck ./...

secrets:
	gitleaks detect --no-git --redact

## test: race detector + atomic coverage profile
test:
	$(GO) test -race -covermode=atomic -coverprofile=$(COVERPROFILE) $(PKG)

## coverage: report coverage and fail below COVERAGE_MIN%
coverage: test
	$(GO) tool cover -func=$(COVERPROFILE)
	$(GO) tool cover -html=$(COVERPROFILE) -o coverage.html
	@total=$$($(GO) tool cover -func=$(COVERPROFILE) | awk '/^total:/ {print $$3}' | tr -d '%'); \
	if awk -v t="$$total" -v m="$(COVERAGE_MIN)" 'BEGIN { exit !(t+0 >= m+0) }'; then \
	  printf 'coverage %s%% meets the %s%% minimum\n' "$$total" "$(COVERAGE_MIN)"; \
	else \
	  printf 'FAIL: coverage %s%% is below the %s%% minimum\n' "$$total" "$(COVERAGE_MIN)"; exit 1; \
	fi

## fuzz: short fuzz run of the deserializer/lexer (skips until Fuzz* tests exist)
fuzz:
	@if grep -rql '^func Fuzz' --include='*_test.go' .; then \
	  $(GO) test -run='^$$' -fuzz=Fuzz -fuzztime=15s .; \
	else \
	  echo "no Fuzz* tests yet — see CLAUDE.md (fuzz the deserializer/lexer)"; \
	fi

## bench: run benchmarks
bench:
	$(GO) test -run='^$$' -bench=. -benchmem $(PKG)

## tidy: ensure go.mod/go.sum are tidy and verified
tidy:
	$(GO) mod tidy
	$(GO) mod verify

## tools: install pinned dev tooling into GOPATH/bin
tools:
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	$(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULN_VERSION)
	$(GO) install github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION)

## hooks: install git pre-commit/pre-push hooks (lefthook)
hooks:
	lefthook install

## clean: remove coverage artifacts
clean:
	rm -f $(COVERPROFILE) coverage.html
