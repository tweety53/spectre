package parse

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/tweety53/spectre/internal/model"
)

var _reqRe = regexp.MustCompile(`^- (R(\d+)): (.*)$`)

// SpecFile reads one capability file into a model.Spec.
func SpecFile(path string) (model.Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.Spec{}, err
	}

	spec := model.Spec{
		Capability: strings.TrimSuffix(filepath.Base(path), ".md"),
		Path:       path,
		Raw:        raw,
	}

	var purpose []string
	section := ""
	line := 0
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line++
		t := sc.Text()

		if strings.HasPrefix(t, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(t, "## "))
			continue
		}

		switch section {
		case "Purpose":
			purpose = append(purpose, t)
		case "Requirements":
			if m := _reqRe.FindStringSubmatch(t); m != nil {
				// The regex only captures \d+, so Atoi fails only on
				// overflow; num then stays 0, which the sequencing checks
				// in check.SpecFindings report as out of sequence.
				num, _ := strconv.Atoi(m[2])
				spec.Reqs = append(spec.Reqs, model.Requirement{
					ID:   m[1],
					Num:  num,
					Text: m[3],
					Refs: Refs(t, line),
					Line: line,
				})
				continue
			}
			if strings.HasPrefix(t, "  ") && len(spec.Reqs) > 0 {
				note := strings.TrimPrefix(t, "  ")
				if strings.TrimSpace(note) != "" {
					last := &spec.Reqs[len(spec.Reqs)-1]
					last.Notes = append(last.Notes, note)
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return model.Spec{}, err
	}
	spec.Purpose = strings.TrimSpace(strings.Join(purpose, "\n"))
	return spec, nil
}
