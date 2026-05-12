package main

import (
	"testing"
)

func TestRemoveStartingSlash(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "no leading slash", raw: "foo/bar", want: "foo/bar"},
		{name: "leading slash", raw: "/foo/bar", want: "foo/bar"},
		{name: "empty string", raw: "", want: ""},
		{name: "only slash", raw: "/", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := funcMap["removeStartingSlash"].(func(string) string)(tt.raw)
			if got != tt.want {
				t.Errorf("removeStartingSlash(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
