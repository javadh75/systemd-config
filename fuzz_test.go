package systemdconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// FuzzDeserialize checks that Deserialize never panics on arbitrary input
// and that, whenever parsing succeeds, the serialized form is canonical:
// parsing and serializing it again reproduces it byte for byte.
func FuzzDeserialize(f *testing.F) {
	seeds := []string{
		"",
		"[Unit]\nDescription=Test\n",
		"[Unit]\r\nDescription=Test\r\n",
		"[Address]\nAddress=10.0.0.1/24\n\n[Address]\nAddress=10.0.0.2/24\n",
		"[Service]\nExecStart=/bin/foo\\\n--bar\n",
		"[Service]\nExecStart=/bin/foo \\\n# comment\n--bar\n",
		"# comment\n; comment\n",
		"[Unit] garbage\n",
		"[Unit\n",
		"Option=before section\n[Unit]\nA=B\n",
		"[Unit]\nEnvironment=FOO=bar\n",
		"[Unit]\nA=\n=B\n",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}

	// every real-world fixture (and its golden form) is a seed too
	fixtures, err := filepath.Glob(filepath.Join("testdata", "*"))
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range fixtures {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // testdata/fuzz is a directory
		}
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		unit, err := Deserialize(bytes.NewReader(data))
		if err != nil {
			return
		}

		var first bytes.Buffer
		if _, err := unit.WriteTo(&first); err != nil {
			t.Fatalf("WriteTo() error = %v", err)
		}

		reparsed, err := Deserialize(bytes.NewReader(first.Bytes()))
		if err != nil {
			t.Fatalf("reparse of canonical output failed: %v\noutput: %q", err, first.String())
		}

		var second bytes.Buffer
		if _, err := reparsed.WriteTo(&second); err != nil {
			t.Fatalf("WriteTo() of reparsed unit error = %v", err)
		}
		if !bytes.Equal(first.Bytes(), second.Bytes()) {
			t.Errorf("canonical form is not a fixpoint:\nfirst:  %q\nsecond: %q", first.String(), second.String())
		}
	})
}
