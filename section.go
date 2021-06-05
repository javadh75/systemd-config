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
