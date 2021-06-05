package systemdconfig

// Unit represents a systemd config unit file.
type Unit struct {
	Sections []*Section
}

// NewUnit returns a new systemd config unit file.
func NewUnit() *Unit {
	return &Unit{Sections: []*Section{}}
}
