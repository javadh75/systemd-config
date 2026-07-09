package systemdconfig

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain fails the suite if any test leaks a goroutine, guarding the
// deserializer against regressing to a design that leaks lexer goroutines
// on early-error returns.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
