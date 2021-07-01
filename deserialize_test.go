package systemdconfig

import (
	"bufio"
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
			got, _, _ := NewLexer(strings.NewReader(tt.args.s))
			buf := new(strings.Builder)

			_, err := io.Copy(buf, got.buf)
			if err != nil {
				t.Errorf("Failed to read got.buf %v", got.buf)
			}
			if buf.String() != tt.args.s {
				t.Errorf("NewLexer() got.buf = %v, want %v", buf.String(), tt.args.s)
			}
		})
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
