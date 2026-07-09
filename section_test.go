package systemdconfig

import (
	"reflect"
	"testing"
)

func TestNewSection(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Section
	}{
		{
			"SimpleSection",
			args{name: "Sec"},
			&Section{Name: "Sec", Options: []*OptionValue{}},
		},
		{
			"EmptySection",
			args{name: ""},
			&Section{Name: "", Options: []*OptionValue{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSection(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSection_OptionHelpers(t *testing.T) {
	s := NewSection("Network")

	if _, ok := s.Value("DNS"); ok {
		t.Error("Value() on empty section reported present")
	}
	if got := s.Values("DNS"); got != nil {
		t.Errorf("Values() on empty section = %v, want nil", got)
	}

	s.AddOption("DNS", "1.1.1.1")
	s.AddOption("Gateway", "10.0.0.1")
	s.AddOption("DNS", "8.8.8.8")

	if got, ok := s.Value("DNS"); !ok || got != "8.8.8.8" {
		t.Errorf("Value(DNS) = %q, %v; want last-wins 8.8.8.8, true", got, ok)
	}
	if got := s.Values("DNS"); !reflect.DeepEqual(got, []string{"1.1.1.1", "8.8.8.8"}) {
		t.Errorf("Values(DNS) = %v, want both values in order", got)
	}
	if len(s.Options) != 3 {
		t.Errorf("len(Options) = %d, want 3", len(s.Options))
	}
}

func TestSection_Match(t *testing.T) {
	type fields struct {
		Name    string
		Options []*OptionValue
	}
	type args struct {
		other *Section
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "DiffrentNames",
			fields: fields{
				Name:    "A",
				Options: []*OptionValue{},
			},
			args: args{
				other: &Section{
					Name:    "B",
					Options: []*OptionValue{},
				},
			},
			want: false,
		},
		{
			name: "DiffrentLengthOfOptions",
			fields: fields{
				Name:    "A",
				Options: []*OptionValue{},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "B",
							Value:  "C",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "DiffrentOptionsWithSameLength1",
			fields: fields{
				Name: "A",
				Options: []*OptionValue{
					{
						Option: "B",
						Value:  "C",
					},
				},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "B",
							Value:  "D",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "DiffrentOptionsWithSameLength2",
			fields: fields{
				Name: "A",
				Options: []*OptionValue{
					{
						Option: "B",
						Value:  "C",
					},
					{
						Option: "D",
						Value:  "C",
					},
				},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "B",
							Value:  "C",
						},
						{
							Option: "D",
							Value:  "E",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "OrderedEqual",
			fields: fields{
				Name: "A",
				Options: []*OptionValue{
					{
						Option: "B",
						Value:  "C",
					},
					{
						Option: "D",
						Value:  "C",
					},
				},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "B",
							Value:  "C",
						},
						{
							Option: "D",
							Value:  "C",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "UnorderedEqual",
			fields: fields{
				Name: "A",
				Options: []*OptionValue{
					{
						Option: "B",
						Value:  "C",
					},
					{
						Option: "D",
						Value:  "C",
					},
				},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "D",
							Value:  "C",
						},
						{
							Option: "B",
							Value:  "C",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "DuplicateOptionsEqual",
			fields: fields{
				Name: "A",
				Options: []*OptionValue{
					{
						Option: "B",
						Value:  "C",
					},
					{
						Option: "B",
						Value:  "C",
					},
				},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "B",
							Value:  "C",
						},
						{
							Option: "B",
							Value:  "C",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "UnorderedDuplicateUnequal",
			fields: fields{
				Name: "A",
				Options: []*OptionValue{
					{
						Option: "B",
						Value:  "C",
					},
					{
						Option: "B",
						Value:  "C",
					},
				},
			},
			args: args{
				other: &Section{
					Name: "A",
					Options: []*OptionValue{
						{
							Option: "D",
							Value:  "C",
						},
						{
							Option: "B",
							Value:  "C",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Section{
				Name:    tt.fields.Name,
				Options: tt.fields.Options,
			}
			if got := s.Match(tt.args.other); got != tt.want {
				t.Errorf("Section.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
