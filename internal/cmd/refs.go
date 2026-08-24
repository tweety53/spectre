package cmd

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/tweety53/spectre/internal/check"
	"github.com/tweety53/spectre/internal/tree"
)

// Refs prints every citation of one requirement, in this tree and in each
// declared peer, then the trees it scanned.
func Refs(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("refs", stderr)
	if err := fs.Parse(args); err != nil {
		return Usage
	}

	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	// The <id> half of the argument follows this tree's own configured id
	// prefix, so the pattern is built only after the tree is resolved.
	targetRe := regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)#(` + regexp.QuoteMeta(t.Cfg.IDPrefix) + `\d+)$`)
	m := targetRe.FindStringSubmatch(fs.Arg(0))
	if fs.NArg() != 1 || m == nil {
		fmt.Fprintf(stderr, "usage: spectre refs <capability>#<id> (this tree's ids start with %q)\n", t.Cfg.IDPrefix)
		return Usage
	}
	capName, reqID := m[1], m[2]

	peers, err := t.Peers()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	specs, err := t.Specs()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	scanned := []string{"this tree"}
	var sources []check.PeerCitationSource
	for _, name := range sortedKeys(peers) {
		rp := tree.ResolvePeer(peers, name)
		switch rp.Resolution {
		case tree.PeerNotPresent:
			scanned = append(scanned, name+" (not present)")
			continue
		case tree.PeerUnreadable:
			scanned = append(scanned, fmt.Sprintf("%s (unreadable: %s)", name, rp.Err))
			continue
		case tree.PeerFound:
			// fall through to processing below.
		default:
			// Defensive: any resolution this switch doesn't know about is
			// treated as unreadable, never as found — rp.Tree is only
			// guaranteed non-nil for PeerFound.
			scanned = append(scanned, fmt.Sprintf("%s (unreadable: %s)", name, rp.Resolution))
			continue
		}
		theirPeers, err := rp.Tree.Peers()
		if err != nil {
			scanned = append(scanned, fmt.Sprintf("%s (unreadable: %s)", name, err))
			continue
		}
		scanned = append(scanned, name)
		sources = append(sources, check.PeerCitationSource{
			Name:     name,
			Root:     rp.Tree.Root,
			Specs:    rp.Specs,
			OurNames: tree.NamesFor(theirPeers, t.Root),
		})
	}

	citations := check.Citations(t.Root, capName, reqID, specs, sources)
	for _, c := range citations {
		prefix := ""
		if c.Peer != "" {
			prefix = c.Peer + ":"
		}
		fmt.Fprintf(stdout, "%s%s:%d: %s cites %s\n", prefix, c.Path, c.Line, c.From, c.Raw)
	}
	if len(citations) == 0 {
		fmt.Fprintf(stdout, "no citations of %s#%s\n", capName, reqID)
	}
	fmt.Fprintf(stdout, "scanned: %s\n", strings.Join(scanned, ", "))
	return OK
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
