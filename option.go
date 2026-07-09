package systemdconfig

import "fmt"

// OptionValue represents an option of a section.
type OptionValue struct {
	Option string
	Value  string
}

// NewOptionValue returns a new OptionValue with pre-set option and value.
func NewOptionValue(option, value string) *OptionValue {
	return &OptionValue{Option: option, Value: value}
}

func (uo *OptionValue) String() string {
	return fmt.Sprintf("{Option: %q, Value: %q}", uo.Option, uo.Value)
}

// Match reports whether uo and other have identical option and value.
func (uo *OptionValue) Match(other *OptionValue) bool {
	return uo.Option == other.Option && uo.Value == other.Value
}
