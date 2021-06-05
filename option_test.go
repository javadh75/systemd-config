package systemdconfig

import (
	"reflect"
	"testing"
)

func TestNewUnitOption(t *testing.T) {
	type args struct {
		option string
		value  string
	}
	tests := []struct {
		name string
		args args
		want *OptionValue
	}{
		{
			"SimpleOption",
			args{option: "Opt", value: "Val"},
			&OptionValue{Option: "Opt", Value: "Val"},
		},
		{
			"EmptyOption",
			args{option: "", value: ""},
			&OptionValue{Option: "", Value: ""},
		},
		{
			"EmptyOpt",
			args{option: "", value: "Val"},
			&OptionValue{Option: "", Value: "Val"},
		},
		{
			"EmptyVal",
			args{option: "Opt", value: ""},
			&OptionValue{Option: "Opt", Value: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUnitOption(tt.args.option, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnitOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptionValue_String(t *testing.T) {
	type fields struct {
		Option string
		Value  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"SimpleOption",
			fields{Option: "Opt", Value: "Val"},
			"{Option: \"Opt\", Value: \"Val\"}",
		},
		{
			"EmptyOption",
			fields{Option: "", Value: ""},
			"{Option: \"\", Value: \"\"}",
		},
		{
			"EmptyOpt",
			fields{Option: "", Value: "Val"},
			"{Option: \"\", Value: \"Val\"}",
		},
		{
			"EmptyVal",
			fields{Option: "Opt", Value: ""},
			"{Option: \"Opt\", Value: \"\"}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uo := &OptionValue{
				Option: tt.fields.Option,
				Value:  tt.fields.Value,
			}
			if got := uo.String(); got != tt.want {
				t.Errorf("OptionValue.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptionValue_Match(t *testing.T) {
	type fields struct {
		Option string
		Value  string
	}
	type args struct {
		other *OptionValue
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"SimpleOptionMatch",
			fields{Option: "Opt", Value: "Val"},
			args{&OptionValue{Option: "Opt", Value: "Val"}},
			true,
		},
		{
			"SimpleOptionMismatch",
			fields{Option: "Opt", Value: "Val"},
			args{&OptionValue{Option: "NotOpt", Value: "NotVal"}},
			false,
		},
		{
			"OptMismatch",
			fields{Option: "Opt", Value: "Val"},
			args{&OptionValue{Option: "NotOpt", Value: "Val"}},
			false,
		},
		{
			"ValueMismatch",
			fields{Option: "Opt", Value: "Val"},
			args{&OptionValue{Option: "Opt", Value: "NotVal"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uo := &OptionValue{
				Option: tt.fields.Option,
				Value:  tt.fields.Value,
			}
			if got := uo.Match(tt.args.other); got != tt.want {
				t.Errorf("OptionValue.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
