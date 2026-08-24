// Package parse turns spectre markdown into the shared model. A Parser
// carries the tree's configuration: a requirement bullet is anchored to
// this tree's own configured id prefix, since it declares a new id in
// this tree's vocabulary. A reference may carry a foreign id-shaped
// token only when it names a peer — RefFindings resolves a
// peer-qualified reference against that peer's own parsed ids, so its
// id is carried as written regardless of this tree's own prefix. A
// reference with no peer names an id in this tree's own vocabulary, so
// it must match this tree's own configured prefix or it is not read as
// a reference at all: otherwise a default tree would start reading
// things like "(@v2)" as citations it never read before.
package parse

import (
	"regexp"

	"github.com/tweety53/spectre/internal/config"
	"github.com/tweety53/spectre/internal/model"
)

// anyID matches any id-shaped token a peer-qualified reference could
// name, independent of which tree's configured prefix produced it:
// config.Load requires every id-prefix to start with a letter and hold
// only letters and hyphens, so this is exactly that shape followed by
// its digits.
const anyID = `[A-Za-z][A-Za-z-]*\d+`

// Parser reads spectre markdown under one tree's configuration.
type Parser struct {
	cfg     config.Config
	reqRe   *regexp.Regexp
	refRe   *regexp.Regexp
	localID *regexp.Regexp
}

// New builds a parser for cfg.
func New(cfg config.Config) *Parser {
	prefix := regexp.QuoteMeta(cfg.IDPrefix)
	return &Parser{
		cfg:     cfg,
		reqRe:   regexp.MustCompile(`^- (` + prefix + `(\d+)): (.*)$`),
		refRe:   regexp.MustCompile(`\(@(?:([a-z0-9][a-z0-9-]*):)?(?:([a-z0-9][a-z0-9-]*)#)?(` + anyID + `)\)`),
		localID: regexp.MustCompile(`^` + prefix + `\d+$`),
	}
}

// Refs extracts every reference in text, tagging each with line. A
// reference with no peer must match this tree's own configured id
// prefix to be read as a reference at all; a peer-qualified reference
// may carry any id-shaped token, since it names an id declared by
// whichever tree the peer resolves to.
func (p *Parser) Refs(text string, line int) []model.Ref {
	var out []model.Ref
	for _, m := range p.refRe.FindAllStringSubmatch(text, -1) {
		if m[1] == "" && !p.localID.MatchString(m[3]) {
			continue
		}
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
