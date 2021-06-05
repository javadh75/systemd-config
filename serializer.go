package systemdconfig

import (
	"bytes"
	"io"
)

// Serializer serialize the given systemd config unit file.
func Serializer(unit *Unit) io.Reader {
	var buf bytes.Buffer

	if len(unit.Sections) == 0 {
		return &buf
	}

	for i, section := range unit.Sections {

		if len(section.Options) == 0 {
			continue
		}

		WriteSectionHeader(&buf, section)
		for _, option := range section.Options {
			WriteOptionValue(&buf, option)
		}
		if i < len(unit.Sections)-1 {
			WriteNewLine(&buf)
		}
	}

	return &buf
}

// WriteNewLine writes a new line to given buffer.
func WriteNewLine(buf *bytes.Buffer) {
	buf.WriteRune('\n')
}

// WriteSectionHeader writes a section header to given buffer.
func WriteSectionHeader(buf *bytes.Buffer, section *Section) {
	buf.WriteRune('[')
	buf.WriteString(section.Name)
	buf.WriteRune(']')
	WriteNewLine(buf)
}

// WriteOptionValue writes a option and value to given buffer.
func WriteOptionValue(buf *bytes.Buffer, option *OptionValue) {
	buf.WriteString(option.Option)
	buf.WriteRune('=')
	buf.WriteString(option.Value)
	WriteNewLine(buf)
}
