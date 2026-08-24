package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tweety53/spectre/internal/tree"
)

var targetRe = regexp.MustCompile(`^([a-z0-9][a-z0-9-]*)#(R\d+)$`)

// Refs prints every citation of one requirement, in this tree and in each
// declared peer, then the trees it scanned.
func Refs(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("refs", stderr)
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	m := targetRe.FindStringSubmatch(fs.Arg(0))
	if fs.NArg() != 1 || m == nil {
		fmt.Fprintln(stderr, "usage: spectre refs <capability>#<id>")
		return Usage
	}
	capName, reqID := m[1], m[2]

	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	peers, err := t.Peers()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	hits := 0
	report := func(prefix, path string, line int, from, raw string) {
		hits++
		fmt.Fprintf(stdout, "%s%s:%d: %s cites %s\n", prefix, path, line, from, raw)
	}

	// This tree: a citation matches when it names no peer and resolves to
	// the target capability, either explicitly or by being same-file.
	specs, err := t.Specs()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	for _, s := range specs {
		for _, r := range s.Reqs {
			for _, ref := range r.Refs {
				if ref.Peer != "" || ref.ID != reqID {
					continue
				}
				if ref.Capability == capName || (ref.Capability == "" && s.Capability == capName) {
					rp, _ := filepath.Rel(t.Root, s.Path)
					report("", rp, ref.Line, r.ID, ref.Raw)
				}
			}
		}
	}

	scanned := []string{"this tree"}
	for _, name := range sortedKeys(peers) {
		path := peers[name]
		if _, err := os.Stat(path); err != nil {
			scanned = append(scanned, name+" (not present)")
			continue
		}
		scanned = append(scanned, name)
		pt, err := tree.Open(path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		theirPeers, err := pt.Peers()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		// Which name does this peer use for us?
		ourNames := map[string]bool{}
		for theirName, theirPath := range theirPeers {
			if sameDir(theirPath, t.Root) {
				ourNames[theirName] = true
			}
		}
		pSpecs, err := pt.Specs()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		for _, s := range pSpecs {
			for _, r := range s.Reqs {
				for _, ref := range r.Refs {
					if ourNames[ref.Peer] && ref.Capability == capName && ref.ID == reqID {
						rp, _ := filepath.Rel(pt.Root, s.Path)
						report(name+":", rp, ref.Line, r.ID, ref.Raw)
					}
				}
			}
		}
	}

	if hits == 0 {
		fmt.Fprintf(stdout, "no citations of %s#%s\n", capName, reqID)
	}
	fmt.Fprintf(stdout, "scanned: %s\n", strings.Join(scanned, ", "))
	return OK
}

func sameDir(a, b string) bool {
	ra, erra := filepath.EvalSymlinks(a)
	rb, errb := filepath.EvalSymlinks(b)
	if erra != nil || errb != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
