package systemdconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRoundTrip deserializes every fixture under testdata/, serializes it
// back, and compares against the committed .golden file. It also checks
// that the canonical (golden) form is a fixpoint: parsing and serializing
// it again must reproduce it byte for byte.
func TestRoundTrip(t *testing.T) {
	goldens, err := filepath.Glob(filepath.Join("testdata", "*.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if len(goldens) == 0 {
		t.Fatal("no .golden files found in testdata")
	}

	for _, golden := range goldens {
		source := strings.TrimSuffix(golden, ".golden")
		t.Run(filepath.Base(source), func(t *testing.T) {
			src, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}

			unit, err := Deserialize(bytes.NewReader(src))
			if err != nil {
				t.Fatalf("Deserialize(%s) error = %v", source, err)
			}
			var out bytes.Buffer
			if _, err := unit.WriteTo(&out); err != nil {
				t.Fatalf("WriteTo() error = %v", err)
			}
			if out.String() != string(want) {
				t.Errorf("serialized %s differs from golden:\ngot:\n%s\nwant:\n%s", source, out.String(), want)
			}

			reparsed, err := Deserialize(bytes.NewReader(out.Bytes()))
			if err != nil {
				t.Fatalf("Deserialize of canonical output error = %v", err)
			}
			var again bytes.Buffer
			if _, err := reparsed.WriteTo(&again); err != nil {
				t.Fatalf("WriteTo() of reparsed unit error = %v", err)
			}
			if again.String() != out.String() {
				t.Errorf("canonical form is not a fixpoint:\nfirst:\n%s\nsecond:\n%s", out.String(), again.String())
			}
		})
	}
}
