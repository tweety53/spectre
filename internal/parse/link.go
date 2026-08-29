package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/tweety53/spectre/internal/model"
)

// The five sections link.md recognises. Each may appear at most once; a
// repeated heading is a parse error, the way peers already errors on a
// repeated name (internal/tree/tree.go's Peers).
const (
	sectionPartOf     = "Part of"
	sectionParts      = "Parts"
	sectionBranch     = "Branch"
	sectionMergeOrder = "Merge order"
	sectionTasksHere  = "Tasks here"
)

var (
	// linkPeerNameRe is the same character rule spectre already applies to
	// a peer name: internal/parse/ref.go's citation grammar accepts
	// exactly this shape for the peer half of "(@peer:capability#id)".
	linkPeerNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

	// linkCodeSpanRe matches a line that is exactly one backtick-delimited
	// span and nothing else.
	linkCodeSpanRe = regexp.MustCompile("^`([^`]*)`$")

	// linkRefSplitRe splits a code span's content into its peer and
	// change-id halves at the first colon.
	linkRefSplitRe = regexp.MustCompile(`^([^:]+):(.+)$`)

	// linkBranchRe is the git ref-name shape
	// <agents repo>/scripts/resolve-base-branch.sh enforces, copied rather
	// than depended on: spectre has no dependency on agents.
	linkBranchRe = regexp.MustCompile(`^[A-Za-z0-9._][A-Za-z0-9._/-]*$`)

	// linkMergeItemRe matches one "## Merge order" line: a list number,
	// then a code span naming "." or a peer.
	linkMergeItemRe = regexp.MustCompile("^[0-9]+\\.\\s+`([^`]*)`$")

	// linkTaskTokenRe matches one "## Tasks here" comma-separated item: a
	// bare number or an ascending "lo-hi" range.
	linkTaskTokenRe = regexp.MustCompile(`^([0-9]+)(?:-([0-9]+))?$`)
)

// linkLine is one non-blank line collected under a recognised section,
// carrying its 1-based source line for error messages.
type linkLine struct {
	text string
	line int
}

// LinkFile reads a change's link.md: the five-section grammar recording a
// change that spans more than one repository's spectre tree
// (design.md's no-free-text-in-link-md). Parsing only — no peer
// resolution, no findings, and no filesystem access beyond this one file,
// the same idiom SpecFile and TasksFile already use. Task numbering
// aside, this grammar has nothing tied to p's configuration.
func (p *Parser) LinkFile(path string) (model.Link, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.Link{}, err
	}

	headingLine := map[string]int{}   // recognised section name -> its heading's line
	bodies := map[string][]linkLine{} // recognised section name -> its non-blank body lines

	section := ""
	line := 0
	inFence := false
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line++
		t := sc.Text()

		// Fence-aware, the same convention spec.go and tasks.go use: a
		// heading-shaped line inside a fenced illustrative example must
		// not be read as a real section.
		if strings.HasPrefix(strings.TrimSpace(t), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		if strings.HasPrefix(t, "## ") {
			name := strings.TrimSpace(strings.TrimPrefix(t, "## "))
			if isLinkSection(name) {
				if first, dup := headingLine[name]; dup {
					return model.Link{}, fmt.Errorf(
						"link.md:%d: duplicate %q section (first seen on line %d)", line, name, first)
				}
				headingLine[name] = line
			}
			section = name
			continue
		}

		if !isLinkSection(section) || strings.TrimSpace(t) == "" {
			continue
		}
		bodies[section] = append(bodies[section], linkLine{text: t, line: line})
	}
	if err := sc.Err(); err != nil {
		return model.Link{}, err
	}

	if err := checkLinkSectionCombinations(headingLine); err != nil {
		return model.Link{}, err
	}

	var link model.Link

	if head, ok := headingLine[sectionPartOf]; ok {
		ref, err := parseSingleLinkRef(sectionPartOf, bodies[sectionPartOf], head)
		if err != nil {
			return model.Link{}, err
		}
		link.PartOf = ref
	}

	if head, ok := headingLine[sectionParts]; ok {
		lines := bodies[sectionParts]
		if len(lines) == 0 {
			return model.Link{}, fmt.Errorf(
				"link.md:%d: %q must list at least one part", head, sectionParts)
		}
		for _, l := range lines {
			ref, err := parseLinkRefLine(l)
			if err != nil {
				return model.Link{}, err
			}
			link.Parts = append(link.Parts, ref)
		}
	}

	if head, ok := headingLine[sectionBranch]; ok {
		lines := bodies[sectionBranch]
		if len(lines) != 1 {
			return model.Link{}, fmt.Errorf(
				"link.md:%d: %q must contain exactly one line", head, sectionBranch)
		}
		b := strings.TrimSpace(lines[0].text)
		if !linkBranchRe.MatchString(b) {
			return model.Link{}, fmt.Errorf(
				"link.md:%d: %q is not a valid branch name", lines[0].line, b)
		}
		link.Branch = b
	}

	if head, ok := headingLine[sectionMergeOrder]; ok {
		lines := bodies[sectionMergeOrder]
		if len(lines) == 0 {
			return model.Link{}, fmt.Errorf(
				"link.md:%d: %q must list at least one entry", head, sectionMergeOrder)
		}
		for _, l := range lines {
			m := linkMergeItemRe.FindStringSubmatch(l.text)
			if m == nil {
				return model.Link{}, fmt.Errorf(
					"link.md:%d: %q is not a numbered \".\" or peer entry", l.line, l.text)
			}
			item := m[1]
			if item != "." && !linkPeerNameRe.MatchString(item) {
				return model.Link{}, fmt.Errorf(
					"link.md:%d: %q is not \".\" or a valid peer name", l.line, item)
			}
			link.MergeOrder = append(link.MergeOrder, item)
		}
	}

	if head, ok := headingLine[sectionTasksHere]; ok {
		lines := bodies[sectionTasksHere]
		if len(lines) != 1 {
			return model.Link{}, fmt.Errorf(
				"link.md:%d: %q must contain exactly one line", head, sectionTasksHere)
		}
		nums, err := parseTaskRanges(lines[0])
		if err != nil {
			return model.Link{}, err
		}
		link.TasksHere = nums
	}

	return link, nil
}

func isLinkSection(name string) bool {
	switch name {
	case sectionPartOf, sectionParts, sectionBranch, sectionMergeOrder, sectionTasksHere:
		return true
	default:
		return false
	}
}

// checkLinkSectionCombinations enforces which sections may appear
// together: links are one level deep, so a satellite's "## Part of"
// section never shares a file with the canonical-only sections, and a
// canonical's "## Parts" section never shares one with the satellite-only
// "## Tasks here".
func checkLinkSectionCombinations(headingLine map[string]int) error {
	_, hasPartOf := headingLine[sectionPartOf]
	_, hasParts := headingLine[sectionParts]
	_, hasMergeOrder := headingLine[sectionMergeOrder]
	_, hasTasksHere := headingLine[sectionTasksHere]

	if hasPartOf && hasParts {
		return fmt.Errorf(
			"link.md: %q and %q cannot both be present; links are one level deep",
			"## "+sectionPartOf, "## "+sectionParts)
	}
	if hasPartOf && hasMergeOrder {
		return fmt.Errorf("link.md: %q does not belong alongside %q",
			"## "+sectionMergeOrder, "## "+sectionPartOf)
	}
	if hasParts && hasTasksHere {
		return fmt.Errorf("link.md: %q does not belong alongside %q",
			"## "+sectionTasksHere, "## "+sectionParts)
	}
	return nil
}

// parseSingleLinkRef parses a section whose body must be exactly one
// "<peer>:<change-id>" code span line ("## Part of"). head is the
// section's own heading line, used when the body is empty.
func parseSingleLinkRef(section string, lines []linkLine, head int) (model.LinkRef, error) {
	if len(lines) != 1 {
		return model.LinkRef{}, fmt.Errorf(
			"link.md:%d: %q must contain exactly one \"<peer>:<change-id>\" reference", head, section)
	}
	return parseLinkRefLine(lines[0])
}

// parseLinkRefLine parses one "<peer>:<change-id>" code span line, used by
// both "## Part of" and "## Parts".
func parseLinkRefLine(l linkLine) (model.LinkRef, error) {
	m := linkCodeSpanRe.FindStringSubmatch(strings.TrimSpace(l.text))
	if m == nil {
		return model.LinkRef{}, fmt.Errorf(
			"link.md:%d: %q must be a single \"<peer>:<change-id>\" code span", l.line, l.text)
	}
	sm := linkRefSplitRe.FindStringSubmatch(m[1])
	if sm == nil {
		return model.LinkRef{}, fmt.Errorf(
			"link.md:%d: %q is not \"<peer>:<change-id>\"", l.line, m[1])
	}
	peer, id := sm[1], sm[2]
	if !linkPeerNameRe.MatchString(peer) {
		return model.LinkRef{}, fmt.Errorf(
			"link.md:%d: %q is not a valid peer name", l.line, peer)
	}
	if !isValidLinkChangeID(id) {
		return model.LinkRef{}, fmt.Errorf(
			"link.md:%d: %q is not a valid change id", l.line, id)
	}
	return model.LinkRef{Peer: peer, ChangeID: id}, nil
}

// parseTaskRanges parses "## Tasks here": comma-separated positive
// integers and ascending ranges, themselves ascending overall with no
// overlaps, expanded into individual task numbers.
func parseTaskRanges(l linkLine) ([]int, error) {
	text := strings.TrimSpace(l.text)
	if text == "" {
		return nil, fmt.Errorf("link.md:%d: %q must list at least one task", l.line, sectionTasksHere)
	}

	var out []int
	max := 0 // highest task number claimed so far; every task number is positive
	for _, tok := range strings.Split(text, ",") {
		lo, hi, err := parseTaskToken(l.line, tok)
		if err != nil {
			return nil, err
		}
		if lo <= max {
			return nil, fmt.Errorf(
				"link.md:%d: %q is out of order or overlaps an earlier item", l.line, tok)
		}
		for n := lo; n <= hi; n++ {
			out = append(out, n)
		}
		max = hi
	}
	return out, nil
}

// parseTaskToken parses one comma-separated item of "## Tasks here": a
// bare positive integer, or a "lo-hi" range with lo strictly less than hi.
func parseTaskToken(line int, tok string) (lo, hi int, err error) {
	m := linkTaskTokenRe.FindStringSubmatch(tok)
	if m == nil {
		return 0, 0, fmt.Errorf("link.md:%d: %q is not a task number or range", line, tok)
	}
	lo, convErr := strconv.Atoi(m[1])
	if convErr != nil || lo == 0 {
		return 0, 0, fmt.Errorf("link.md:%d: %q must be a positive integer", line, tok)
	}
	if m[2] == "" {
		return lo, lo, nil
	}
	hi, convErr = strconv.Atoi(m[2])
	if convErr != nil || hi == 0 {
		return 0, 0, fmt.Errorf("link.md:%d: %q must be a positive integer", line, tok)
	}
	if lo >= hi {
		return 0, 0, fmt.Errorf("link.md:%d: range %q must be ascending", line, tok)
	}
	return lo, hi, nil
}

// isValidLinkChangeID mirrors internal/cmd's isValidID: a change id must
// be usable both as a directory name and as a required heading. It is
// reimplemented here rather than imported because cmd depends on tree,
// which depends on parse — parse importing cmd would be a cycle. Keep the
// two in agreement by hand if either changes.
func isValidLinkChangeID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	if strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return false
	}
	if filepath.Clean(id) != id {
		return false
	}
	if filepath.IsAbs(id) {
		return false
	}
	if strings.ContainsFunc(id, unicode.IsControl) {
		return false
	}
	// A backtick would close the code span a link.md entry embeds a change
	// id in, corrupting the rendered file. Reject it here so no id ever
	// reaches that renderer un-escaped.
	if strings.Contains(id, "`") {
		return false
	}
	return true
}
