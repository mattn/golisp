package golisp

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "",
			want:  "()",
		},
		{
			input: "1",
			want:  "(1)",
		},
		{
			input: "(1)(2)",
			want:  "((1) (2))",
		},
		{
			input: "(1)(2)",
			want:  "((1) (2))",
		},
		{
			input: "1 2",
			want:  "(1 2)",
		},
		{
			input: "t",
			want:  "(t)",
		},
		{
			input: "nil",
			want:  "(nil)",
		},
	}
	for _, test := range tests {
		t.Logf("%q", test.input)
		parser := NewParser(strings.NewReader(test.input))
		node, err := parser.ParseParen()
		if err != nil {
			t.Error(err)
		}
		got := node.String()

		if got != test.want {
			t.Errorf("want %q for %q but got %q", test.want, test.input, got)
		}
	}
}
