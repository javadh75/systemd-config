package systemdconfig

// Merge returns the effective configuration of a unit file combined with
// its drop-ins, following systemd.unit(5) semantics: the dropins are
// applied after base, in the order given, as if their content were
// appended to it. Assigning an option the empty value resets it — every
// earlier occurrence of that option in sections of the same name is
// removed, and the empty assignment itself is dropped (systemd treats an
// empty assignment as "back to default"). Sections are never collapsed:
// duplicate sections (e.g. [Address] in .network files) stay separate
// and addressable, and sections left empty by a reset are kept.
//
// The inputs are not modified; the result shares no memory with them.
// Use Unit.Value and Unit.Values on the result to read effective values.
func Merge(base *Unit, dropins ...*Unit) *Unit {
	merged := &Unit{}
	for _, u := range append([]*Unit{base}, dropins...) {
		for _, s := range u.Sections {
			section := &Section{Name: s.Name, Options: []*OptionValue{}}
			merged.Sections = append(merged.Sections, section)
			for _, o := range s.Options {
				if o.Value == "" {
					resetOption(merged, s.Name, o.Option)
					continue
				}
				section.Options = append(section.Options, &OptionValue{Option: o.Option, Value: o.Value})
			}
		}
	}
	return merged
}

// resetOption removes every occurrence of the named option from all
// sections of the unit with the given name.
func resetOption(u *Unit, section, option string) {
	for _, s := range u.Sections {
		if s.Name != section {
			continue
		}
		kept := s.Options[:0]
		for _, o := range s.Options {
			if o.Option != option {
				kept = append(kept, o)
			}
		}
		s.Options = kept
	}
}
