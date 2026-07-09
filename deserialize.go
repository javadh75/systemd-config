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
)

type lexDataType int

const (
	sectionKind lexDataType = iota
	optionKind
)

type lexData struct {
	Type    lexDataType
	Option  *OptionValue
	Section *Section
}

type lexer struct {
	buf     *bufio.Reader
	lexchan chan *lexData
	errchan chan error
}

type lexStep func() (lexStep, error)

// newLexer returns a new systemd config lexer and needed data and error channel
func newLexer(f io.Reader) (*lexer, <-chan *lexData, <-chan error) {
	lexchan := make(chan *lexData)
	errchan := make(chan error, 1)
	buf := bufio.NewReader(f)

	return &lexer{buf: buf, lexchan: lexchan, errchan: errchan}, lexchan, errchan
}

func (l *lexer) Lex() {
	defer func() {
		close(l.lexchan)
		close(l.errchan)
	}()
	next := l.LexNextSection
	for next != nil {
		var err error
		next, err = next()
		if err != nil {
			l.errchan <- err
			return
		}
	}
}

func (l *lexer) LexNextSection() (lexStep, error) {
	r, _, err := l.buf.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = nil
		}
		return nil, err
	}

	if r == '[' {
		return l.LexSectionName, nil
	} else if IsComment(r) {
		return l.IgnoreLineFunc(l.LexNextSection), nil
	}
	return l.LexNextSection, nil
}

func (l *lexer) LexSectionName() (lexStep, error) {
	sec, err := l.buf.ReadBytes(']')
	if err != nil {
		return nil, errors.New("unable to find end of section")
	}

	return l.LexSectionSuffixFunc(string(sec[:len(sec)-1])), nil
}

func (l *lexer) LexSectionSuffixFunc(section string) lexStep {
	return func() (lexStep, error) {
		garbage, _, err := l.toEOL()
		if err != nil {
			return nil, err
		}

		garbage = bytes.TrimSpace(garbage)
		if len(garbage) > 0 {
			return nil, fmt.Errorf("found garbage after section name %s: %q", section, garbage)
		}

		l.lexchan <- &lexData{
			Type:    sectionKind,
			Section: &Section{Name: section, Options: []*OptionValue{}},
			Option:  nil,
		}

		return l.LexNextSectionOrOptionFunc(section), nil
	}
}

func (l *lexer) LexNextSectionOrOptionFunc(section string) lexStep {
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

func (l *lexer) LexOptionNameFunc(section string) lexStep {
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

func (l *lexer) LexOptionValueFunc(section, name string) lexStep {
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
			// mirroring systemd.syntax(7)
			if bytes.HasSuffix(line, []byte{'\\'}) && !eof {
				partial.Write(line[:len(line)-1])
				partial.WriteRune(' ')
				continue
			}

			partial.Write(line)
			break
		}

		l.lexchan <- &lexData{
			Type:    optionKind,
			Section: nil,
			Option:  &OptionValue{Option: name, Value: strings.TrimSpace(partial.String())},
		}

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

// Deserialize parses the given systemd config into a Unit.
func Deserialize(f io.Reader) (*Unit, error) {
	lexer, lexchan, errchan := newLexer(f)

	go lexer.Lex()

	unit := Unit{}

	for ld := range lexchan {
		switch ld.Type {
		case optionKind:
			if ld.Option != nil {
				if len(unit.Sections) == 0 {
					return nil, errors.New("unit file misparse: option before section")
				}

				s := len(unit.Sections) - 1
				unit.Sections[s].Options = append(unit.Sections[s].Options, ld.Option)
			}
		case sectionKind:
			if ld.Section != nil {
				unit.Sections = append(unit.Sections, ld.Section)
			}
		}
	}

	err := <-errchan

	return &unit, err
}
