package cmd

import "testing"

// TestExitCodeContract pins the numeric exit codes every command relies on.
// Every assertion in this package compares an exit code against these
// constants by name, so nothing catches OK/Fail/Usage's numeric values
// drifting — changing Usage from 2 to 3 leaves the whole suite green. This
// test is the one place that pins the actual numbers.
func TestExitCodeContract(t *testing.T) {
	if OK != 0 {
		t.Errorf("OK = %d, want 0", OK)
	}
	if Fail != 1 {
		t.Errorf("Fail = %d, want 1", Fail)
	}
	if Usage != 2 {
		t.Errorf("Usage = %d, want 2", Usage)
	}
}
