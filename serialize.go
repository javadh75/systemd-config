package systemdconfig

import (
	"bytes"
	"fmt"
	"io"
)

// Serialize serializes the given systemd config unit file.
func Serialize(unit *Unit) io.Reader {
	var buf bytes.Buffer
	_, _ = unit.WriteTo(&buf)
	return &buf
}

// WriteTo writes the serialized unit to w, implementing io.WriterTo.
func (u *Unit) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer

	for i, section := range u.Sections {
		writeSectionHeader(&buf, section)
		for _, option := range section.Options {
			writeOptionValue(&buf, option)
		}
		if i < len(u.Sections)-1 {
			writeNewLine(&buf)
		}
	}

	n, err := buf.WriteTo(w)
	if err != nil {
		return n, fmt.Errorf("writing unit: %w", err)
	}
	return n, nil
}

// writeNewLine writes a new line to given buffer.
func writeNewLine(buf *bytes.Buffer) {
	buf.WriteRune('\n')
}

// writeSectionHeader writes a section header to given buffer.
func writeSectionHeader(buf *bytes.Buffer, section *Section) {
	buf.WriteRune('[')
	buf.WriteString(section.Name)
	buf.WriteRune(']')
	writeNewLine(buf)
}

// writeOptionValue writes an option and value to given buffer.
func writeOptionValue(buf *bytes.Buffer, option *OptionValue) {
	buf.WriteString(option.Option)
	buf.WriteRune('=')
	buf.WriteString(option.Value)
	writeNewLine(buf)
}
