package systemdconfig

import (
	"reflect"
	"testing"
)

func TestNewUnit(t *testing.T) {
	tests := []struct {
		name string
		want *Unit
	}{
		{
			"SimpleUnit",
			&Unit{Sections: []*Section{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUnit(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnit_SectionLookupAndAdd(t *testing.T) {
	u := NewUnit()
	if got := u.SectionByName("Address"); got != nil {
		t.Errorf("SectionByName() on empty unit = %v, want nil", got)
	}
	if got := u.SectionsByName("Address"); got != nil {
		t.Errorf("SectionsByName() on empty unit = %v, want nil", got)
	}

	first := u.AddSection("Address")
	u.AddSection("Route")
	second := u.AddSection("Address")

	if got := u.SectionByName("Address"); got != first {
		t.Errorf("SectionByName() = %v, want first Address section", got)
	}
	if got := u.SectionsByName("Address"); !reflect.DeepEqual(got, []*Section{first, second}) {
		t.Errorf("SectionsByName() = %v, want both Address sections in order", got)
	}
	if len(u.Sections) != 3 {
		t.Errorf("len(Sections) = %d, want 3", len(u.Sections))
	}
}

func TestUnit_Value(t *testing.T) {
	unit := unitOf(
		sectionOf("Network",
			optionOf("DHCP", "no"),
			optionOf("DNS", "192.168.0.1"),
		),
		sectionOf("Address",
			optionOf("Address", "10.0.0.1/24"),
		),
		sectionOf("Network",
			optionOf("DNS", "1.1.1.1"),
		),
	)

	tests := []struct {
		name    string
		section string
		option  string
		want    string
		wantOK  bool
	}{
		{"LastDuplicateSectionWins", "Network", "DNS", "1.1.1.1", true},
		{"FallsBackToEarlierDuplicate", "Network", "DHCP", "no", true},
		{"SingleSection", "Address", "Address", "10.0.0.1/24", true},
		{"MissingOption", "Network", "Gateway", "", false},
		{"MissingSection", "Route", "Gateway", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := unit.Value(tt.section, tt.option)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Unit.Value(%q, %q) = %q, %v, want %q, %v",
					tt.section, tt.option, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestUnit_String(t *testing.T) {
	unit := unitOf(
		sectionOf("Unit", optionOf("Description", "Test")),
		sectionOf("Install", optionOf("WantedBy", "multi-user.target")),
	)
	want := "[Unit]\nDescription=Test\n\n[Install]\nWantedBy=multi-user.target\n"
	if got := unit.String(); got != want {
		t.Errorf("Unit.String() = %q, want %q", got, want)
	}
}

func TestUnit_Match(t *testing.T) {
	type fields struct {
		Sections []*Section
	}
	type args struct {
		other *Unit
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "UnEqualSectionLength",
			fields: fields{
				Sections: []*Section{
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{},
				},
			},
			want: false,
		},
		{
			name: "EqualEmptySections",
			fields: fields{
				Sections: []*Section{},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{},
				},
			},
			want: true,
		},
		{
			name: "UnequalSectionName",
			fields: fields{
				Sections: []*Section{
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{
						{
							Name:    "B",
							Options: []*OptionValue{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "UnequalMultipleSections",
			fields: fields{
				Sections: []*Section{
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
					{
						Name:    "B",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{
						{
							Name:    "A",
							Options: []*OptionValue{},
						},
						{
							Name:    "C",
							Options: []*OptionValue{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "UnequalMultipleSectionsWithOptions",
			fields: fields{
				Sections: []*Section{
					{
						Name: "A",
						Options: []*OptionValue{
							{
								Option: "AA",
								Value:  "AV",
							},
							{
								Option: "AB",
								Value:  "AV",
							},
						},
					},
					{
						Name:    "B",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{
						{
							Name: "A",
							Options: []*OptionValue{
								{
									Option: "AA",
									Value:  "AV",
								},
								{
									Option: "AB",
									Value:  "AV",
								},
							},
						},
						{
							Name:    "C",
							Options: []*OptionValue{},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "EqualMultipleSectionsWithOptions",
			fields: fields{
				Sections: []*Section{
					{
						Name: "A",
						Options: []*OptionValue{
							{
								Option: "AA",
								Value:  "AV",
							},
							{
								Option: "AB",
								Value:  "AV",
							},
						},
					},
					{
						Name:    "B",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{
						{
							Name: "A",
							Options: []*OptionValue{
								{
									Option: "AA",
									Value:  "AV",
								},
								{
									Option: "AB",
									Value:  "AV",
								},
							},
						},
						{
							Name:    "B",
							Options: []*OptionValue{},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "EqualDuplicateSections",
			fields: fields{
				Sections: []*Section{
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{
						{
							Name:    "A",
							Options: []*OptionValue{},
						},
						{
							Name:    "A",
							Options: []*OptionValue{},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "UnequalMultipleDuplicateSections",
			fields: fields{
				Sections: []*Section{
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
					{
						Name:    "A",
						Options: []*OptionValue{},
					},
				},
			},
			args: args{
				other: &Unit{
					Sections: []*Section{
						{
							Name:    "A",
							Options: []*OptionValue{},
						},
						{
							Name:    "C",
							Options: []*OptionValue{},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Unit{
				Sections: tt.fields.Sections,
			}
			if got := u.Match(tt.args.other); got != tt.want {
				t.Errorf("Unit.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
