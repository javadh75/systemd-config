package systemdconfig

// Section represents a section and its options.
type Section struct {
	Name    string
	Options []*OptionValue
}

// NewSection returns a new section with pre-set name and empty options.
func NewSection(name string) *Section {
	return &Section{Name: name, Options: []*OptionValue{}}
}

// AddOption appends a new option with the given name and value and
// returns it.
func (s *Section) AddOption(option, value string) *OptionValue {
	o := NewOptionValue(option, value)
	s.Options = append(s.Options, o)
	return o
}

// Value returns the value of the last occurrence of the named option,
// following systemd's last-assignment-wins rule, and whether the option
// is present at all.
func (s *Section) Value(option string) (string, bool) {
	for i := len(s.Options) - 1; i >= 0; i-- {
		if s.Options[i].Option == option {
			return s.Options[i].Value, true
		}
	}
	return "", false
}

// Values returns the values of every occurrence of the named option, in
// order of appearance. It returns nil when the option is absent.
func (s *Section) Values(option string) []string {
	var values []string
	for _, o := range s.Options {
		if o.Option == option {
			values = append(values, o.Value)
		}
	}
	return values
}

// Match reports whether s and other have the same name and the same
// options, regardless of option order. Duplicate options must appear the
// same number of times in both sections.
func (s *Section) Match(other *Section) bool {
	if s.Name != other.Name {
		return false
	}
	if len(s.Options) != len(other.Options) {
		return false
	}

	otherSeen := make([]bool, len(other.Options))
	for _, sElement := range s.Options {
		matched := false
		for j, otherElement := range other.Options {
			if !otherSeen[j] && sElement.Match(otherElement) {
				otherSeen[j] = true
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
