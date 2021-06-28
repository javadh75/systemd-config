package systemdconfig

// Unit represents a systemd config unit file.
type Unit struct {
	Sections []*Section
}

// NewUnit returns a new systemd config unit file.
func NewUnit() *Unit {
	return &Unit{Sections: []*Section{}}
}

func (u *Unit) Match(other *Unit) bool {
	if len(u.Sections) != len(other.Sections) {
		return false
	}

	otherSeen := InitialCompareSliceGenerator(len(other.Sections))

	compareList := InitialCompareSliceGenerator(len(u.Sections))

	for i, uElement := range u.Sections {
		for j, otherElement := range other.Sections {
			if uElement.Match(otherElement) {
				compareList[i] = true
				otherSeen[j] = true
				break
			} else if j == len(other.Sections)-1 {
				return false
			}
		}
	}
	if !AllAreTrue(otherSeen) || !AllAreTrue(compareList) {
		return false
	}
	return true
}
