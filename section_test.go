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
