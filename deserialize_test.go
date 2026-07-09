package systemdconfig

import (
	"bufio"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestIsComment(t *testing.T) {
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Comment ;",
			args: args{
				r: ';',
			},
			want: true,
		},
		{
			name: "Comment #",
			args: args{
				r: '#',
			},
			want: true,
		},
		{
			name: "NonComment",
			args: args{
				r: '"',
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsComment(tt.args.r); got != tt.want {
				t.Errorf("isComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLexer(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "EmptyInput",
			args: args{
				s: "",
			},
		},
		{
			name: "SimpleInput",
			args: args{
				s: "ABCDE",
			},
		},
		{
			name: "NotSimpleInput",
			args: args{
				s: "AB\nC\rD\nE",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, _ := newLexer(strings.NewReader(tt.args.s))
			buf := new(strings.Builder)

			_, err := io.Copy(buf, got.buf)
			if err != nil {
				t.Errorf("Failed to read got.buf %v", got.buf)
			}
			if buf.String() != tt.args.s {
				t.Errorf("newLexer() got.buf = %v, want %v", buf.String(), tt.args.s)
			}
		})
	}
}

func TestDeserialize(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *Unit
		wantErr bool
	}{
		{
			name: "Empty",
			in:   "",
			want: &Unit{},
		},
		{
			name: "SimpleSection",
			in:   "[Unit]\nDescription=Test\n",
			want: &Unit{Sections: []*Section{
				{Name: "Unit", Options: []*OptionValue{{Option: "Description", Value: "Test"}}},
			}},
		},
		{
			name: "MSDOSLineEndings",
			in:   "[Unit]\r\nDescription=Test\r\n",
			want: &Unit{Sections: []*Section{
				{Name: "Unit", Options: []*OptionValue{{Option: "Description", Value: "Test"}}},
			}},
		},
		{
			name: "CommentsAndBlankLines",
			in:   "# leading comment\n\n[Unit]\n; another comment\nDescription=Test\n",
			want: &Unit{Sections: []*Section{
				{Name: "Unit", Options: []*OptionValue{{Option: "Description", Value: "Test"}}},
			}},
		},
		{
			name: "DuplicateSections",
			in:   "[Address]\nAddress=10.0.0.1/24\n\n[Address]\nAddress=10.0.0.2/24\n",
			want: &Unit{Sections: []*Section{
				{Name: "Address", Options: []*OptionValue{{Option: "Address", Value: "10.0.0.1/24"}}},
				{Name: "Address", Options: []*OptionValue{{Option: "Address", Value: "10.0.0.2/24"}}},
			}},
		},
		{
			name: "OptionNameAndValueTrimmed",
			in:   "[Unit]\nDescription = Test value \n",
			want: &Unit{Sections: []*Section{
				{Name: "Unit", Options: []*OptionValue{{Option: "Description", Value: "Test value"}}},
			}},
		},
		{
			name: "ContinuationJoinedWithSpace",
			in:   "[Service]\nExecStart=/bin/foo \\\n--bar\n",
			want: &Unit{Sections: []*Section{
				{Name: "Service", Options: []*OptionValue{{Option: "ExecStart", Value: "/bin/foo  --bar"}}},
			}},
		},
		{
			name: "ContinuationSkipsCommentLines",
			in:   "[Service]\nExecStart=/bin/foo \\\n# interleaved comment\n--bar\n",
			want: &Unit{Sections: []*Section{
				{Name: "Service", Options: []*OptionValue{{Option: "ExecStart", Value: "/bin/foo  --bar"}}},
			}},
		},
		{
			name: "TrailingBackslashAtEOFDropped",
			in:   "[Service]\nExecStart=/bin/foo \\",
			want: &Unit{Sections: []*Section{
				{Name: "Service", Options: []*OptionValue{{Option: "ExecStart", Value: "/bin/foo"}}},
			}},
		},
		{
			name:    "GarbageAfterSectionName",
			in:      "[Unit] junk\nDescription=Test\n",
			wantErr: true,
		},
		{
			name:    "UnterminatedSectionName",
			in:      "[Unit\nDescription=Test\n",
			wantErr: true,
		},
		{
			name:    "NewlineInOptionName",
			in:      "[Unit]\nDesc\nription=Test\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Deserialize(strings.NewReader(tt.in))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Deserialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeserializeLineTooLong(t *testing.T) {
	in := "[Unit]\nDescription=" + strings.Repeat("x", LineMax+1) + "\n"
	_, err := Deserialize(strings.NewReader(in))
	if !errors.Is(err, ErrLineTooLong) {
		t.Errorf("Deserialize() error = %v, want ErrLineTooLong", err)
	}
}

func Test_lexer_toEOL(t *testing.T) {
	type fields struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		want1   bool
		wantErr bool
	}{
		{
			name: "EmptyLine",
			fields: fields{
				s: "\n",
			},
			want:    []byte{},
			want1:   false,
			wantErr: false,
		},
		{
			name: "EmptyLineMS-DOS",
			fields: fields{
				s: "\r\n",
			},
			want:    []byte{},
			want1:   false,
			wantErr: false,
		},
		{
			name: "SimpleLine",
			fields: fields{
				s: "SimpleLine\n",
			},
			want:    []byte("SimpleLine"),
			want1:   false,
			wantErr: false,
		},
		{
			name: "EOF",
			fields: fields{
				s: "",
			},
			want:    []byte(""),
			want1:   true,
			wantErr: false,
		},
		// TODO Add test case to cover not EOF error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				buf:     bufio.NewReader(strings.NewReader(tt.fields.s)),
				lexchan: make(chan *lexData),
				errchan: make(chan error, 1),
			}
			got, got1, err := l.toEOL()
			if (err != nil) != tt.wantErr {
				t.Errorf("lexer.toEOL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("lexer.toEOL() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("lexer.toEOL() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
