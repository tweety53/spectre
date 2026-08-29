package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tweety53/spectre/internal/check"
	"github.com/tweety53/spectre/internal/model"
	"github.com/tweety53/spectre/internal/parse"
	"github.com/tweety53/spectre/internal/tree"
)

// linkRefRow is a model.LinkRef with the JSON shape design.md's two-way-link
// decision fixes for a list row: "peer" and "changeId".
type linkRefRow struct {
	Peer     string `json:"peer"`
	ChangeID string `json:"changeId"`
}

type changeRow struct {
	ID    string `json:"id"`
	Done  int    `json:"done"`
	Total int    `json:"total"`

	// PartOf and Parts are set only for a change carrying a link.md — a
	// satellite gets PartOf, a canonical gets Parts, and a change with
	// neither link carries neither field (omitempty), per the JSON
	// contract design.md's two-way-link decision states.
	PartOf *linkRefRow  `json:"partOf,omitempty"`
	Parts  []linkRefRow `json:"parts,omitempty"`

	// ProgressUnknown is true only for a satellite whose canonical
	// progress could not be resolved (peer-absence-is-not-a-finding).
	// Marked explicitly (design.md's list-degrades-and-says-so) rather
	// than left for a consumer to infer from Done/Total being 0: /flow's
	// implementation phase reads "total == 0" off spectre list --json as
	// "no plan spectre can read" and would otherwise send a change whose
	// plan is intact in the canonical repository back to brainstorming
	// just because its peer is not checked out here. omitempty keeps
	// every row without an unresolved link — which is every row today —
	// byte-identical to before this field existed.
	ProgressUnknown bool `json:"progressUnknown,omitempty"`
}

type specRow struct {
	Capability   string `json:"capability"`
	Requirements int    `json:"requirements"`
}

// List prints open changes with task progress, or capabilities under --specs.
func List(args []string, stdout, stderr io.Writer) int {
	fs, root := flagSet("list", stderr)
	specs := fs.Bool("specs", false, "list capabilities instead of changes")
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return Usage
	}
	t, err := resolve(*root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	if *specs {
		found, err := t.Specs()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return Usage
		}
		rows := make([]specRow, 0, len(found))
		for _, s := range found {
			rows = append(rows, specRow{Capability: s.Capability, Requirements: len(s.Reqs)})
		}
		if *asJSON {
			return writeJSON(stdout, stderr, map[string]any{"specs": rows})
		}
		for _, r := range rows {
			fmt.Fprintf(stdout, "%s  %d requirements\n", r.Capability, r.Requirements)
		}
		return OK
	}

	changes, err := t.Changes()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}

	// Resolved unconditionally, the way Validate resolves peers: a
	// satellite's progress needs its declared peer resolved regardless of
	// whether --json is set, and resolving only when a link.md is seen
	// would need a second pass over changes. A malformed peers file is a
	// content problem (design.md's list-degrades-and-says-so), not a
	// reason for a read-only listing to abort: every row still renders,
	// with no peer resolved for any link it carries.
	declared, err := t.Peers()
	if err != nil {
		declared = map[string]string{}
	}
	peers := resolvePeers(declared)

	rows := make([]changeRow, 0, len(changes))
	for _, c := range changes {
		rows = append(rows, linkRow(t, peers, c))
	}
	if *asJSON {
		return writeJSON(stdout, stderr, map[string]any{"changes": rows})
	}
	for _, r := range rows {
		switch {
		case r.PartOf != nil && r.ProgressUnknown:
			fmt.Fprintf(stdout, "%s  ->  %s  -\n", r.ID, r.PartOf.Peer)
		case r.PartOf != nil:
			fmt.Fprintf(stdout, "%s  ->  %s  %d/%d\n", r.ID, r.PartOf.Peer, r.Done, r.Total)
		default:
			fmt.Fprintf(stdout, "%s  %d/%d\n", r.ID, r.Done, r.Total)
		}
	}
	return OK
}

// linkRow builds c's row, reading its link.md when present. A change with
// no link.md gets a plain row. A satellite ("## Part of") gets PartOf and
// its progress resolved from the canonical change in the declared peer,
// per design.md's two-way-link and peer-absence-is-not-a-finding: a peer
// that is not declared, not present, or a counterpart that cannot be
// found, leaves ProgressUnknown true rather than erroring or omitting the
// row. A canonical ("## Parts") gets Parts, verbatim from link.md — no
// peer resolution needed, since Parts is only ever restated, not counted.
//
// A link.md that fails to parse is the same kind of content problem a
// malformed peers file is (design.md's list-degrades-and-says-so):
// validate already reports it as a finding, so linkRow renders this one
// change's row as if it carried no link.md at all rather than aborting
// the whole listing over it.
func linkRow(t *tree.Tree, peers map[string]tree.ResolvedPeer, c model.Change) changeRow {
	row := changeRow{ID: c.ID, Done: c.DoneCount(), Total: len(c.Tasks)}

	linkPath := filepath.Join(c.Dir, check.LinkFile)
	if !fileExists(linkPath) {
		return row
	}
	link, err := parse.New(t.Cfg).LinkFile(linkPath)
	if err != nil {
		return row
	}

	if link.PartOf != (model.LinkRef{}) {
		row.PartOf = &linkRefRow{Peer: link.PartOf.Peer, ChangeID: link.PartOf.ChangeID}
		row.ProgressUnknown = true
		if done, total, ok := linkProgress(peers, link.PartOf); ok {
			row.Done, row.Total = done, total
			row.ProgressUnknown = false
		}
	}
	for _, part := range link.Parts {
		row.Parts = append(row.Parts, linkRefRow{Peer: part.Peer, ChangeID: part.ChangeID})
	}
	return row
}

// linkProgress resolves ref's canonical task progress: rp must be declared
// and PeerFound, and the counterpart change (searched changes/<id> then
// changes/archive/<id>, independent-archive) must exist and read cleanly.
// Any other outcome returns ok == false — peer-absence-is-not-a-finding
// applies to list the same as it does to validate, so no case here is an
// error, only "not known".
func linkProgress(peers map[string]tree.ResolvedPeer, ref model.LinkRef) (done, total int, ok bool) {
	rp, declared := peers[ref.Peer]
	if !declared || rp.Resolution != tree.PeerFound {
		return 0, 0, false
	}
	dir := filepath.Join(rp.Tree.ChangesDir(), ref.ChangeID)
	if !fileExists(filepath.Join(dir, tree.TasksFile)) {
		dir = filepath.Join(rp.Tree.ArchiveDir(), ref.ChangeID)
	}
	tasks, err := parse.New(rp.Tree.Cfg).TasksFile(filepath.Join(dir, tree.TasksFile))
	if err != nil {
		return 0, 0, false
	}
	for _, tk := range tasks {
		if tk.Done {
			done++
		}
	}
	return done, len(tasks), true
}

// fileExists reports whether path names a regular, readable file.
func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func writeJSON(stdout, stderr io.Writer, v any) int {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	return OK
}
