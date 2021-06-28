package systemdconfig

import (
	"io"
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
		f io.Reader
		s string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "EmptyInput",
			args: args{
				f: strings.NewReader(""),
				s: "",
			},
		},
		{
			name: "SimpleInput",
			args: args{
				f: strings.NewReader("ABCDE"),
				s: "ABCDE",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, _ := NewLexer(tt.args.f)
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
