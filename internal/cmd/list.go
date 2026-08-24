package cmd

import (
	"encoding/json"
	"fmt"
	"io"
)

type changeRow struct {
	ID    string `json:"id"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
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

	changes, err := t.Changes(false)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return Usage
	}
	rows := make([]changeRow, 0, len(changes))
	for _, c := range changes {
		rows = append(rows, changeRow{ID: c.ID, Done: c.DoneCount(), Total: len(c.Tasks)})
	}
	if *asJSON {
		return writeJSON(stdout, stderr, map[string]any{"changes": rows})
	}
	for _, r := range rows {
		fmt.Fprintf(stdout, "%s  %d/%d\n", r.ID, r.Done, r.Total)
	}
	return OK
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
