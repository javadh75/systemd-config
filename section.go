package systemdconfig

// Section represents a setion and it's options.
type Section struct {
	Name    string
	Options []*OptionValue
}

// NewSection returns a new section with pre-set name and empty options.
func NewSection(name string) *Section {
	return &Section{Name: name, Options: []*OptionValue{}}
}

// initialCompareSliceGenerator returns an all-false slice used to track
// which elements have already been matched during comparison.
func initialCompareSliceGenerator(size int) []bool {
	ICS := make([]bool, size)
	for index := range ICS {
		ICS[index] = false
	}
	return ICS
}

// allAreTrue reports whether every element of b is true.
func allAreTrue(b []bool) bool {
	for _, element := range b {
		if !element {
			return false
		}
	}
	return true
}

// Match reports whether s and other have the same name and the same
// options, regardless of option order.
func (s *Section) Match(other *Section) bool {
	if s.Name != other.Name {
		return false
	}
	if len(s.Options) != len(other.Options) {
		return false
	}

	otherSeen := initialCompareSliceGenerator(len(other.Options))

	compareList := initialCompareSliceGenerator(len(s.Options))

	for i, sElement := range s.Options {
		for j, otherElement := range other.Options {
			if sElement.Match(otherElement) {
				compareList[i] = true
				otherSeen[j] = true
				break
			} else if j == len(other.Options)-1 {
				return false
			}
		}
	}
	if !allAreTrue(otherSeen) || !allAreTrue(compareList) {
		return false
	}
	return true
}
