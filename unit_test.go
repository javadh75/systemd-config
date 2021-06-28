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
