// Package render turns the model back into spectre markdown. Every write
// in spectre goes through here, so malformed output is unrepresentable.
package render

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tweety53/spectre/internal/model"
)

// Spec renders one capability file.
func Spec(s model.Spec) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# %s\n\n## Purpose\n%s\n\n## Requirements\n", collapseNewlines(s.Capability), s.Purpose)
	for _, r := range s.Reqs {
		fmt.Fprintf(&b, "- %s: %s\n", r.ID, collapseNewlines(r.Text))
		for _, n := range r.Notes {
			for _, line := range strings.Split(n, "\n") {
				if strings.TrimSpace(line) != "" {
					fmt.Fprintf(&b, "  %s\n", line)
				}
			}
		}
	}
	return b.Bytes()
}

// collapseNewlines replaces all newlines with single spaces.
func collapseNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

// Tasks renders a change's tasks.md, renumbering from 1 in slice order.
func Tasks(ts []model.Task) []byte {
	var b bytes.Buffer
	b.WriteString("# Tasks\n\n")
	for i, t := range ts {
		box := " "
		if t.Done {
			box = "x"
		}
		fmt.Fprintf(&b, "- [%s] %d. %s\n", box, i+1, t.Text)
	}
	return b.Bytes()
}
