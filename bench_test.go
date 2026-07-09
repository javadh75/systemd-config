package systemdconfig

import (
	"io"
	"strings"
	"testing"
)

var benchInput = strings.Repeat("[Section]\nKeyOne=some value\nKeyTwo=other value\n\n", 64)

func BenchmarkDeserialize(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := Deserialize(strings.NewReader(benchInput)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSerialize(b *testing.B) {
	unit, err := Deserialize(strings.NewReader(benchInput))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := unit.WriteTo(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
