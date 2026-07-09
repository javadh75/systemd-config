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
