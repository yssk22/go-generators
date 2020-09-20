package helper

import (
	"testing"
)

func TestStrings_ToSnakeCase(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			input:  "FooBarBaz",
			output: "foo_bar_baz",
		},
		{
			input:  "MyURL",
			output: "my_url",
		},
		{
			input:  "URLIsNotGood",
			output: "url_is_not_good",
		},
		{
			input:  "URL123IsNotGood",
			output: "url123_is_not_good",
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(tt *testing.T) {
			got := ToSnakeCase(c.input)
			if got != c.output {
				tt.Errorf("expected: %s, got: %s", c.output, got)
			}
		})
	}
}
