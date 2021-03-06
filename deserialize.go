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
	// SYSTEMD_LINE_MAX mimics the maximum line length that systemd can use.
	// On typical systemd platforms (i.e. modern Linux), this will most
	// commonly be 2048, so let's use that as a sanity check.
	// Technically, we should probably pull this at runtime:
	//    SYSTEMD_LINE_MAX = int(C.sysconf(C.__SC_LINE_MAX))
	// but this would introduce an (unfortunate) dependency on cgo
	SYSTEMD_LINE_MAX = 2048

	// SYSTEMD_NEWLINE defines characters that systemd considers indicators
	// for a newline.
	SYSTEMD_NEWLINE = "\r\n"
)

var (
	// ErrLineTooLong gets returned when a line is too long for systemd to handle.
	ErrLineTooLong = fmt.Errorf("line too long (max %d bytes)", SYSTEMD_LINE_MAX)
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

// NewLexer returns a new systemd config lexer and needed data and error channel
func NewLexer(f io.Reader) (*lexer, <-chan *lexData, <-chan error) {
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
		if l.buf.Buffered() >= SYSTEMD_LINE_MAX {
			line, err := l.buf.Peek(SYSTEMD_LINE_MAX)
			if err != nil {
				l.errchan <- err
				return
			}
			if !bytes.ContainsAny(line, SYSTEMD_NEWLINE) {
				l.errchan <- ErrLineTooLong
				return
			}
		}
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
		if err == io.EOF {
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
			if err == io.EOF {
				err = nil
			}
			return nil, err
		}

		if unicode.IsSpace(r) {
			return l.LexNextSectionOrOptionFunc(section), nil
		} else if r == '[' {
			return l.LexSectionName, nil
		} else if IsComment(r) {
			return l.IgnoreLineFunc(l.LexNextSectionOrOptionFunc(section)), nil
		}

		l.buf.UnreadRune()
		return l.LexOptionNameFunc(section), nil
	}
}

func (l *lexer) LexOptionNameFunc(section string) lexStep {
	return func() (lexStep, error) {
		var partial bytes.Buffer
		for {
			r, _, err := l.buf.ReadRune()
			if err != nil {
				return nil, err
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
		return l.LexOptionValueFunc(section, name, bytes.Buffer{}), nil
	}
}

func (l *lexer) LexOptionValueFunc(section, name string, partial bytes.Buffer) lexStep {
	return func() (lexStep, error) {
		for {
			line, eof, err := l.toEOL()
			if err != nil {
				return nil, err
			}

			if len(bytes.TrimSpace(line)) == 0 {
				break
			}

			partial.Write(line)

			// lack of continuation means this value has been exhausted
			idx := bytes.LastIndex(line, []byte{'\\'})
			if idx == -1 || idx != (len(line)-1) {
				break
			}

			if !eof {
				partial.WriteRune('\n')
			}

			return l.LexOptionValueFunc(section, name, partial), nil
		}

		val := partial.String()
		if strings.HasSuffix(val, "\n") {
			// A newline was added to the end, so the file didn't end with a backslash.
			// => Keep the newline
			val = strings.TrimSpace(val) + "\n"
		} else {
			val = strings.TrimSpace(val)
		}
		l.lexchan <- &lexData{
			Type:    optionKind,
			Section: nil,
			Option:  &OptionValue{Option: name, Value: val},
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
	if err != nil && err != io.EOF {
		return nil, false, err
	}

	line = bytes.TrimSuffix(line, []byte{'\n'})
	line = bytes.TrimSuffix(line, []byte{'\r'})

	return line, err == io.EOF, nil
}

func IsComment(r rune) bool {
	return r == '#' || r == ';'
}

// Deserialize deserialize given unparsed systemd config.
func Deserialize(f io.Reader) (*Unit, error) {
	lexer, lexchan, errchan := NewLexer(f)

	go lexer.Lex()

	unit := Unit{}

	for ld := range lexchan {
		switch ld.Type {
		case optionKind:
			if ld.Option != nil {
				if len(unit.Sections) == 0 {
					return nil, fmt.Errorf("Unit file misparse: option before section")
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
