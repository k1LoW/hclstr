package fmt

import (
	"os"
	"testing"

	"github.com/tenntenn/golden"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name    string
		fmtcmds map[string]string
		in      string
	}{
		{
			"format inline JSON",
			map[string]string{
				"policy": "cat ? | jq . > ?.tmp && mv ?.tmp ?",
			},
			"../testdata/jq_fmt.tf",
		},
		{
			"format inline JSON with prettier",
			map[string]string{
				"policy": "prettier ? --write --parser json",
			},
			"../testdata/prettier_fmt.tf",
		},
		{
			"skip string with directive",
			map[string]string{
				"block": "exit 1",
			},
			"../testdata/directive_fmt.hcl",
		},
		{
			"skip string with function",
			map[string]string{
				"block": "prettier ? --write --parser json",
			},
			"../testdata/with_func_fmt.hcl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.fmtcmds)
			b, err := os.ReadFile(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			formatted, err := c.Format(b)
			if err != nil {
				t.Fatal(err)
			}
			got := string(formatted)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, "", tt.in, got)
				return
			}
			if diff := golden.Diff(t, "", tt.in, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
