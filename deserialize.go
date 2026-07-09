package systemdconfig

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	// LineMax mimics LONG_LINE_MAX, the maximum length of a single line
	// modern systemd accepts in a unit file (1 MiB). Older systemd used
	// LINE_MAX (2048).
	LineMax = 1024 * 1024
)

var (
	// ErrLineTooLong gets returned when a line is too long for systemd to handle.
	ErrLineTooLong = fmt.Errorf("line too long (max %d bytes)", LineMax)

	// ErrAssignmentOutsideSection gets returned when an assignment (or any
	// other non-comment text) appears before the first section header.
	// systemd rejects such lines too.
	ErrAssignmentOutsideSection = errors.New("assignment outside of section")
)

type lexer struct {
	buf  *bufio.Reader
	unit *Unit
}

type lexStep func() (lexStep, error)

// newLexer returns a lexer that parses f into a fresh unit.
func newLexer(f io.Reader) *lexer {
	return &lexer{buf: bufio.NewReader(f), unit: &Unit{}}
}

// lex drives the state machine until the input is exhausted or a step fails.
func (l *lexer) lex() error {
	next := l.LexNextSection
	for next != nil {
		var err error
		next, err = next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lexer) LexNextSection() (lexStep, error) {
	r, _, err := l.buf.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return nil, err
	}

	switch {
	case unicode.IsSpace(r):
		return l.LexNextSection, nil
	case r == '[':
		return l.LexSectionName, nil
	case IsComment(r):
		return l.IgnoreLineFunc(l.LexNextSection), nil
	}
	return nil, ErrAssignmentOutsideSection
}

func (l *lexer) LexSectionName() (lexStep, error) {
	sec, err := l.buf.ReadBytes(']')
	if err != nil {
		return nil, errors.New("unable to find end of section")
	}

	return l.LexSectionSuffixFunc(string(sec[:len(sec)-1])), nil
}

func (l *lexer) LexSectionSuffixFunc(name string) lexStep {
	return func() (lexStep, error) {
		garbage, _, err := l.toEOL()
		if err != nil {
			return nil, err
		}

		garbage = bytes.TrimSpace(garbage)
		if len(garbage) > 0 {
			return nil, fmt.Errorf("found garbage after section name %s: %q", name, garbage)
		}

		section := &Section{Name: name, Options: []*OptionValue{}}
		l.unit.Sections = append(l.unit.Sections, section)

		return l.LexNextSectionOrOptionFunc(section), nil
	}
}

func (l *lexer) LexNextSectionOrOptionFunc(section *Section) lexStep {
	return func() (lexStep, error) {
		r, _, err := l.buf.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return nil, err
		}

		switch {
		case unicode.IsSpace(r):
			return l.LexNextSectionOrOptionFunc(section), nil
		case r == '[':
			return l.LexSectionName, nil
		case IsComment(r):
			return l.IgnoreLineFunc(l.LexNextSectionOrOptionFunc(section)), nil
		}

		if err := l.buf.UnreadRune(); err != nil {
			return nil, fmt.Errorf("unreading rune: %w", err)
		}
		return l.LexOptionNameFunc(section), nil
	}
}

func (l *lexer) LexOptionNameFunc(section *Section) lexStep {
	return func() (lexStep, error) {
		var partial bytes.Buffer
		for {
			r, _, err := l.buf.ReadRune()
			if err != nil {
				return nil, fmt.Errorf("reading option name: %w", err)
			}

			if r == '\n' || r == '\r' {
				return nil, errors.New("unexpected newline encountered while parsing option name")
			}

			if r == '=' {
				break
			}

			partial.WriteRune(r)
		}

		name := strings.TrimSpace(partial.String())
		return l.LexOptionValueFunc(section, name), nil
	}
}

func (l *lexer) LexOptionValueFunc(section *Section, name string) lexStep {
	return func() (lexStep, error) {
		var partial bytes.Buffer
		for {
			line, eof, err := l.toEOL()
			if err != nil {
				return nil, err
			}

			// comment lines inside a continuation are skipped entirely
			if partial.Len() > 0 && len(line) > 0 && IsComment(rune(line[0])) {
				if eof {
					break
				}
				continue
			}

			if len(bytes.TrimSpace(line)) == 0 {
				break
			}

			// a line ending in a backslash is concatenated with the next
			// non-comment line and the backslash is replaced by a space,
			// mirroring systemd.syntax(7); at EOF there is nothing left to
			// concatenate, so the marker simply disappears
			if bytes.HasSuffix(line, []byte{'\\'}) {
				partial.Write(line[:len(line)-1])
				partial.WriteRune(' ')
				if eof {
					break
				}
				continue
			}

			partial.Write(line)
			break
		}

		val := strings.TrimSpace(partial.String())
		// a value ending in a backslash cannot be represented: serializing
		// it would re-trigger line continuation on the next parse, so the
		// dangling marker is dropped
		for strings.HasSuffix(val, `\`) {
			val = strings.TrimSpace(val[:len(val)-1])
		}

		section.Options = append(section.Options, &OptionValue{Option: name, Value: val})

		return l.LexNextSectionOrOptionFunc(section), nil
	}
}

func (l *lexer) IgnoreLineFunc(next lexStep) lexStep {
	return func() (lexStep, error) {
		for {
			line, _, err := l.toEOL()
			if err != nil {
				return nil, err
			}

			line = bytes.TrimSuffix(line, []byte{' '})

			if !bytes.HasSuffix(line, []byte{'\\'}) {
				break
			}
		}

		return next, nil
	}
}

func (l *lexer) toEOL() ([]byte, bool, error) {
	line, err := l.buf.ReadBytes('\n')
	// ignore EOF here since it's roughly equivalent to EOL
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, false, fmt.Errorf("reading line: %w", err)
	}

	line = bytes.TrimSuffix(line, []byte{'\n'})
	line = bytes.TrimSuffix(line, []byte{'\r'})

	if len(line) > LineMax {
		return nil, false, ErrLineTooLong
	}

	return line, errors.Is(err, io.EOF), nil
}

// IsComment reports whether r marks the start of a comment line ('#' or ';').
func IsComment(r rune) bool {
	return r == '#' || r == ';'
}

// Deserialize parses the given systemd config into a Unit. On error it
// returns the sections parsed so far alongside the error.
func Deserialize(f io.Reader) (*Unit, error) {
	l := newLexer(f)
	if err := l.lex(); err != nil {
		return l.unit, err
	}
	return l.unit, nil
}
