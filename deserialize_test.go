package systemdconfig

import (
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

func Test_isLexer(t *testing.T) {
	type args struct {
		t interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nilType",
			args: args{
				t: nil,
			},
			want: false,
		},
		{
			name: "otherType",
			args: args{
				t: 1,
			},
			want: false,
		},
		{
			name: "lexerType",
			args: args{
				t: lexer{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLexer(tt.args.t); got != tt.want {
				t.Errorf("isLexer() = %v, want %v", got, tt.want)
			}
		})
	}
}
