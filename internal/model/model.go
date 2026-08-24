// Package model holds the types every spectre package shares.
package model

// Spec is one capability file under specs/.
type Spec struct {
	Capability string // filename without .md
	Purpose    string // prose under "## Purpose", newlines preserved
	Reqs       []Requirement
	Path       string // absolute path on disk
	Raw        []byte // the file's bytes, as parsed
}

// Requirement is one "- R<n>: ..." bullet and the indented prose beneath it.
type Requirement struct {
	ID    string // "R1"
	Num   int    // 1
	Text  string // bullet text after the id, verbatim
	Notes []string
	Refs  []Ref
	Line  int // 1-based line of the bullet
}

// Ref is a citation of a requirement. Peer is empty for a same-tree
// reference; Capability is empty for a same-file reference.
type Ref struct {
	Peer       string
	Capability string
	ID         string
	Raw        string // "@gymie:plans#R7"
	Line       int
}

// Change is one folder under changes/ or changes/archive/.
type Change struct {
	ID       string
	Dir      string // absolute path
	Archived bool
	Tasks    []Task
}

// Task is one "- [ ] <n>. ..." line in tasks.md.
type Task struct {
	Num  int
	Text string
	Done bool
	Line int
}

// Done reports how many of the change's tasks are checked.
func (c Change) DoneCount() int {
	n := 0
	for _, t := range c.Tasks {
		if t.Done {
			n++
		}
	}
	return n
}
