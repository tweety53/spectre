// Package parse turns spectre markdown into the shared model.
package parse

import (
	"regexp"

	"github.com/tweety53/spectre/internal/model"
)

// (@R4) same file, (@plans#R4) same tree, (@gymie:plans#R4) peer tree.
var _refRe = regexp.MustCompile(`\(@(?:([a-z0-9][a-z0-9-]*):)?(?:([a-z0-9][a-z0-9-]*)#)?(R\d+)\)`)

// Refs extracts every reference in text, tagging each with line.
func Refs(text string, line int) []model.Ref {
	var out []model.Ref
	for _, m := range _refRe.FindAllStringSubmatch(text, -1) {
		out = append(out, model.Ref{
			Peer:       m[1],
			Capability: m[2],
			ID:         m[3],
			Raw:        "@" + trimParens(m[0]),
			Line:       line,
		})
	}
	return out
}

func trimParens(s string) string {
	return s[2 : len(s)-1] // strip "(@" and ")"
}
