// Package render turns the model back into spectre markdown. Every write
// in spectre goes through here, so malformed output is unrepresentable.
package render

import (
	"bytes"
	"fmt"

	"github.com/tweety53/spectre/internal/model"
)

// Spec renders one capability file.
func Spec(s model.Spec) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# %s\n\n## Purpose\n%s\n\n## Requirements\n", s.Capability, s.Purpose)
	for _, r := range s.Reqs {
		fmt.Fprintf(&b, "- %s: %s\n", r.ID, r.Text)
		for _, n := range r.Notes {
			fmt.Fprintf(&b, "  %s\n", n)
		}
	}
	return b.Bytes()
}
