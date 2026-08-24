package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func seedTree(t *testing.T) string {
	t.Helper()
	base := emptyTree(t)
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(base, "spectre", rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("specs/auth.md", "# auth\n\n## Purpose\nP.\n\n## Requirements\n- R1: The system SHALL a.\n- R2: The system SHALL b.\n")
	write("changes/kan-1-first/proposal.md", "# kan-1-first\n\n## Why\nBecause.\n\n## What changes\n- a thing\n")
	write("changes/kan-1-first/tasks.md", "# Tasks\n\n- [x] 1. One\n- [ ] 2. Two\n- [ ] 3. Three\n")
	return base
}

func TestListChanges(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", seedTree(t)}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d, stderr = %s", code, errBuf.String())
	}
	want := "kan-1-first  1/3\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestListSpecs(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", seedTree(t), "--specs"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	want := "auth  2 requirements\n"
	if out.String() != want {
		t.Errorf("stdout = %q, want %q", out.String(), want)
	}
}

func TestListJSON(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", seedTree(t), "--json"}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	var got struct {
		Changes []struct {
			ID    string `json:"id"`
			Done  int    `json:"done"`
			Total int    `json:"total"`
		} `json:"changes"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json: %v (%s)", err, out.String())
	}
	if len(got.Changes) != 1 || got.Changes[0].ID != "kan-1-first" || got.Changes[0].Done != 1 || got.Changes[0].Total != 3 {
		t.Errorf("json = %+v", got.Changes)
	}
}

func TestListEmptyTree(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := List([]string{"--root", emptyTree(t)}, &out, &errBuf); code != OK {
		t.Fatalf("exit = %d", code)
	}
	if out.String() != "" {
		t.Errorf("stdout = %q, want empty", out.String())
	}
}
