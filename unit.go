package systemdconfig

import "bytes"

// Unit represents a systemd config unit file.
type Unit struct {
	Sections []*Section
}

// NewUnit returns a new systemd config unit file.
func NewUnit() *Unit {
	return &Unit{Sections: []*Section{}}
}

// SectionsByName returns all sections with the given name, in order of
// appearance. It returns nil when no section matches.
func (u *Unit) SectionsByName(name string) []*Section {
	var sections []*Section
	for _, s := range u.Sections {
		if s.Name == name {
			sections = append(sections, s)
		}
	}
	return sections
}

// SectionByName returns the first section with the given name, or nil
// when no section matches.
func (u *Unit) SectionByName(name string) *Section {
	for _, s := range u.Sections {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Value returns the value of the named option in the named section and
// whether the option is present at all. It follows systemd's semantics
// for duplicates: sections with the same name behave as one merged
// section and the last assignment wins.
func (u *Unit) Value(section, option string) (string, bool) {
	for i := len(u.Sections) - 1; i >= 0; i-- {
		if u.Sections[i].Name != section {
			continue
		}
		if v, ok := u.Sections[i].Value(option); ok {
			return v, true
		}
	}
	return "", false
}

// String returns the canonical serialized form of the unit,
// implementing fmt.Stringer.
func (u *Unit) String() string {
	var buf bytes.Buffer
	_, _ = u.WriteTo(&buf)
	return buf.String()
}

// AddSection appends a new empty section with the given name and returns it.
func (u *Unit) AddSection(name string) *Section {
	s := NewSection(name)
	u.Sections = append(u.Sections, s)
	return s
}

// Match reports whether u and other contain the same sections, regardless
// of section order. Duplicate sections must appear the same number of
// times in both units.
func (u *Unit) Match(other *Unit) bool {
	if len(u.Sections) != len(other.Sections) {
		return false
	}

	otherSeen := make([]bool, len(other.Sections))
	for _, uElement := range u.Sections {
		matched := false
		for j, otherElement := range other.Sections {
			if !otherSeen[j] && uElement.Match(otherElement) {
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
