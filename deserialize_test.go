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
