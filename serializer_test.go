package systemdconfig

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestSerializer(t *testing.T) {
	type args struct {
		unit *Unit
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"EmptyUnit",
			args{
				unit: &Unit{
					Sections: []*Section{},
				},
			},
			"",
		},
		{
			"EmptyOption",
			args{
				unit: &Unit{
					Sections: []*Section{
						&Section{
							Name:    "Match",
							Options: []*OptionValue{},
						},
					},
				},
			},
			"",
		},
		{
			"SimpleUnit",
			args{
				unit: &Unit{
					Sections: []*Section{
						&Section{
							Name: "Match",
							Options: []*OptionValue{
								&OptionValue{
									Option: "A",
									Value:  "B",
								},
								&OptionValue{
									Option: "C",
									Value:  "D",
								},
							},
						},
					},
				},
			},
			`[Match]
A=B
C=D
`,
		},
		{
			"TwoSection",
			args{
				unit: &Unit{
					Sections: []*Section{
						&Section{
							Name: "AAA",
							Options: []*OptionValue{
								&OptionValue{
									Option: "A",
									Value:  "B",
								},
							},
						},
						&Section{
							Name: "BBB",
							Options: []*OptionValue{
								&OptionValue{
									Option: "A",
									Value:  "B",
								},
							},
						},
					},
				},
			},
			`[AAA]
A=B

[BBB]
A=B
`,
		},
		{
			"TwoDuplicateSection",
			args{
				unit: &Unit{
					Sections: []*Section{
						&Section{
							Name: "AAA",
							Options: []*OptionValue{
								&OptionValue{
									Option: "A",
									Value:  "B",
								},
							},
						},
						&Section{
							Name: "AAA",
							Options: []*OptionValue{
								&OptionValue{
									Option: "A",
									Value:  "B",
								},
							},
						},
					},
				},
			},
			`[AAA]
A=B

[AAA]
A=B
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outReader := Serializer(tt.args.unit)
			outBytes, err := ioutil.ReadAll(outReader)
			if err != nil {
				t.Errorf("case %s: encountered error while reading output: %v", tt.name, err)
			}
			output := string(outBytes)
			if !reflect.DeepEqual(output, tt.want) {
				t.Errorf("Serializer() = %v, want %v", output, tt.want)
			}
		})
	}
}

func TestWriteNewLine(t *testing.T) {
	var buf, want bytes.Buffer
	want.WriteRune('\n')
	t.Run("SimpleWriteNewLine", func(t *testing.T) {
		if WriteNewLine(&buf); !reflect.DeepEqual(buf, want) {
			t.Errorf("WriteNewLine() buf %v, want %v", buf, want)
		}
	})
}

func TestWriteSectionHeader(t *testing.T) {
	var empty_byte_slice []byte
	type args struct {
		buf     *bytes.Buffer
		section *Section
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"SimpleOption",
			args{
				buf: bytes.NewBuffer(empty_byte_slice),
				section: &Section{
					Name:    "Sec",
					Options: []*OptionValue{},
				},
			},
		},
		{
			"EmptySection",
			args{
				buf: bytes.NewBuffer(empty_byte_slice),
				section: &Section{
					Name:    "",
					Options: []*OptionValue{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteRune('[')
			buf.WriteString(tt.args.section.Name)
			buf.WriteRune(']')
			WriteNewLine(&buf)
			if WriteSectionHeader(tt.args.buf, tt.args.section); !reflect.DeepEqual(tt.args.buf.Bytes(), buf.Bytes()) {
				t.Errorf("WriteSectionHeader() given buffer %v, buf %v", tt.args.buf.String(), buf.String())
			}
		})
	}
}

func TestWriteOptionValue(t *testing.T) {
	var empty_byte_slice []byte
	type args struct {
		buf    *bytes.Buffer
		option *OptionValue
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"SimpleOption",
			args{
				buf: bytes.NewBuffer(empty_byte_slice),
				option: &OptionValue{
					Option: "Opt",
					Value:  "Val",
				},
			},
		},
		{
			"EmptyOption",
			args{
				buf: bytes.NewBuffer(empty_byte_slice),
				option: &OptionValue{
					Option: "",
					Value:  "",
				},
			},
		},
		{
			"EmptyValue",
			args{
				buf: bytes.NewBuffer(empty_byte_slice),
				option: &OptionValue{
					Option: "Opt",
					Value:  "",
				},
			},
		},
		{
			"EmptyOpt",
			args{
				buf: bytes.NewBuffer(empty_byte_slice),
				option: &OptionValue{
					Option: "",
					Value:  "Val",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteString(tt.args.option.Option)
			buf.WriteRune('=')
			buf.WriteString(tt.args.option.Value)
			WriteNewLine(&buf)
			if WriteOptionValue(tt.args.buf, tt.args.option); !reflect.DeepEqual(tt.args.buf.Bytes(), buf.Bytes()) {
				t.Errorf("WriteOptionValue() given buffer %v, buf %v", tt.args.buf.String(), buf.String())
			}
		})
	}
}
