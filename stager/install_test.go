package stager

import (
	"testing"

	"github.com/pelletier/go-toml"
)

func TestStringTemplateInterpolation(t *testing.T) {
	tcs := []struct {
		name, input, output string
		tree                map[string]interface{}
	}{
		{"concat", "\"blue\" + \"berry\"", "blueberry", nil},
		{"basic", "something", "else", map[string]interface{}{"something": "else"}},
		{"nested", "a.b", "c", map[string]interface{}{"a": map[string]interface{}{"b": "c"}}},
		{"numbers and spaces", "s + ' aaa'", "22 aaa", map[string]interface{}{"s": 22}},
		{"complex", "\"s/^KERNELRELEASE.*/\"+base.linux.version+\"/g\"", "s/^KERNELRELEASE.*/5.6.19/g", map[string]interface{}{
			"base": map[string]interface{}{"linux": map[string]interface{}{"version": "5.6.19"}}}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tree, _ := toml.TreeFromMap(tc.tree)

			out, err := evalStringSection(tc.input, tree)
			if err != nil {
				t.Errorf("evalStringSection() failed: %v", err)
			}
			if out != tc.output {
				t.Errorf("evalStringSection(%q) returned %q, want %q", tc.input, out, tc.output)
			}
		})
	}
}
