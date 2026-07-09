package systemdconfig

// Unit represents a systemd config unit file.
type Unit struct {
	Sections []*Section
}

// NewUnit returns a new systemd config unit file.
func NewUnit() *Unit {
	return &Unit{Sections: []*Section{}}
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
