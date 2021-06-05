package systemdconfig

import "fmt"

// Repreents a option of an section.
type OptionValue struct {
	Option string
	Value  string
}

// NewUnitOption returns a new UnitOption with pre-set value.
func NewUnitOption(option, value string) *OptionValue {
	return &OptionValue{Option: option, Value: value}
}

func (uo *OptionValue) String() string {
	return fmt.Sprintf("{Option: %q, Value: %q", uo.Option, uo.Value)
}

// Match compares two UnitOptions and returns true if they are identical.
func (uo *OptionValue) Match(other *OptionValue) bool {
	return uo.Option == other.Option && uo.Value == other.Value
}
