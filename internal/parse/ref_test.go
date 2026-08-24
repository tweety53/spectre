package parse

import (
	"testing"

	"github.com/tweety53/spectre/internal/config"
)

// TestRefsIgnoresPeerlessForeignPrefix pins the fix round 1 correction: a
// reference with no peer must match this tree's own configured id prefix
// or it is not read as a reference at all. Without this, a default tree
// (id-prefix "R") would start reading arbitrary parenthesized text like
// "(@v2)" as a citation it never read before anyID was introduced.
func TestRefsIgnoresPeerlessForeignPrefix(t *testing.T) {
	parser := New(config.Default()) // id-prefix "R"
	for _, text := range []string{
		"See (@v2) for details.",
		"See (@user2) for details.",
		"See (@issue3) for details.",
		"See (@auth#Q9) for details.",
	} {
		if refs := parser.Refs(text, 1); len(refs) != 0 {
			t.Errorf("Refs(%q) = %+v, want none: a peerless reference must match the local id prefix", text, refs)
		}
	}
}

// TestRefsPeerQualifiedForeignPrefixStillMatches is the other half of the
// same rule: a peer-qualified reference may still carry an id in the
// peer's own prefix, different from this tree's own.
func TestRefsPeerQualifiedForeignPrefixStillMatches(t *testing.T) {
	parser := New(config.Default()) // id-prefix "R"
	refs := parser.Refs("(@gymie:billing#REQ-1)", 1)
	if len(refs) != 1 {
		t.Fatalf("Refs = %+v, want one peer-qualified ref", refs)
	}
	if ref := refs[0]; ref.Peer != "gymie" || ref.Capability != "billing" || ref.ID != "REQ-1" {
		t.Errorf("ref = %+v, want peer gymie, capability billing, id REQ-1", ref)
	}
}

// TestRefsLocalPrefixStillMatches confirms peerless references in this
// tree's own prefix are unaffected by the tightening.
func TestRefsLocalPrefixStillMatches(t *testing.T) {
	parser := New(config.Default()) // id-prefix "R"
	refs := parser.Refs("(@R1) and (@plans#R2)", 1)
	if len(refs) != 2 {
		t.Fatalf("Refs = %+v, want two local-prefix refs", refs)
	}
	if refs[0].Peer != "" || refs[0].Capability != "" || refs[0].ID != "R1" {
		t.Errorf("same-file ref = %+v", refs[0])
	}
	if refs[1].Peer != "" || refs[1].Capability != "plans" || refs[1].ID != "R2" {
		t.Errorf("same-tree ref = %+v", refs[1])
	}
}

// TestRefsConfiguredLocalPrefixStillMatches confirms the tightening
// checks the parser's own configured prefix, not a hardcoded "R".
func TestRefsConfiguredLocalPrefixStillMatches(t *testing.T) {
	cfg := config.Default()
	cfg.IDPrefix = "REQ-"
	parser := New(cfg)
	if refs := parser.Refs("(@REQ-1)", 1); len(refs) != 1 || refs[0].ID != "REQ-1" {
		t.Errorf("Refs = %+v, want one local ref REQ-1", refs)
	}
	// "R1" is not this tree's configured prefix ("REQ-"), and is
	// peerless, so it must not be read as a reference.
	if refs := parser.Refs("(@R1)", 1); len(refs) != 0 {
		t.Errorf("Refs = %+v, want none: R1 does not match the configured prefix REQ-", refs)
	}
}
