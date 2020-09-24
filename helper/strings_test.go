package helper

import (
	"reflect"
	"testing"
)

func TestStrings_ToLowerCamleCase(t *testing.T) {
	cases := []struct {
		input          string
		useUpperAtHead bool
		output         string
	}{
		{
			input:  "foo_bar_baz",
			output: "fooBarBaz",
		},
		{
			input:  "my_url",
			output: "myUrl",
		},
		{
			input:  "url_is_not_good",
			output: "urlIsNotGood",
		},
		{
			input:  "url123_is_not_good",
			output: "url123IsNotGood",
		},
		{
			input:  "URLIsNotID",
			output: "urlIsNotId",
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(tt *testing.T) {
			got := ToLowerCamleCase(c.input)
			if got != c.output {
				tt.Errorf("expected: %s, got: %s", c.output, got)
			}
		})
	}
}

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

func TestStrings_tokenize(t *testing.T) {
	cases := []struct {
		input  string
		output []string
	}{
		{
			input:  "foo_bar_baz",
			output: []string{"foo", "bar", "baz"},
		},
		{
			input:  "fooBarBaz",
			output: []string{"foo", "Bar", "Baz"},
		},
		{
			input:  "URLIsNotID",
			output: []string{"URL", "Is", "Not", "ID"},
		},
		{
			input:  "URL123IsNotID",
			output: []string{"URL123", "Is", "Not", "ID"},
		},
		{
			input:  "URL123isNotID",
			output: []string{"URL123", "is", "Not", "ID"},
		},
	}
	for _, c := range cases {
		t.Run(c.input, func(tt *testing.T) {
			got := tokenize(c.input)
			if !reflect.DeepEqual(got, c.output) {
				tt.Errorf("expected: %s, got: %s", c.output, got)
			}
		})
	}
}
