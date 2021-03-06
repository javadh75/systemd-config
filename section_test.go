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

func TestInitialCompareSliceGenerator(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want []bool
	}{
		{
			name: "Empty",
			args: args{
				size: 0,
			},
			want: []bool{},
		},
		{
			name: "Simple",
			args: args{
				size: 3,
			},
			want: []bool{false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitialCompareSliceGenerator(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitialCompareSliceGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllAreTrue(t *testing.T) {
	type args struct {
		b []bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Empty",
			args: args{
				b: []bool{},
			},
			want: true,
		},
		{
			name: "AllTrue",
			args: args{
				b: []bool{true, true, true},
			},
			want: true,
		},
		{
			name: "AllFalse",
			args: args{
				b: []bool{false, false, false},
			},
			want: false,
		},
		{
			name: "OneFalse",
			args: args{
				b: []bool{false, true, true},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AllAreTrue(tt.args.b); got != tt.want {
				t.Errorf("AllAreTrue() = %v, want %v", got, tt.want)
			}
		})
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
