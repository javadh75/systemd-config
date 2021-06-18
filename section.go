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

func InitialCompareSliceGenerator(size int) []bool {
	ICS := make([]bool, size)
	for index, _ := range ICS {
		ICS[index] = false
	}
	return ICS
}

func AllAreTrue(b []bool) bool {
	for _, element := range b {
		if !element {
			return false
		}
	}
	return true
}

func (s *Section) Match(other *Section) bool {
	if s.Name != other.Name {
		return false
	}
	if len(s.Options) != len(other.Options) {
		return false
	}

	otherSeen := InitialCompareSliceGenerator(len(other.Options))

	compareList := InitialCompareSliceGenerator(len(s.Options))

	shouldBreak := false
	for i, sElement := range s.Options {
		for j, otherElement := range other.Options {
			if sElement.Match(otherElement) {
				compareList[i] = true
				otherSeen[j] = true
				break
			} else if j == len(other.Options) {
				shouldBreak = true
			}
		}
		if shouldBreak {
			break
		}
	}
	if !AllAreTrue(otherSeen) || !AllAreTrue(compareList) {
		return false
	}
	return true
}
