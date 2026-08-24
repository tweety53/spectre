package parse

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tweety53/spectre/internal/model"
)

var _taskRe = regexp.MustCompile(`^- \[([ x])\] (\d+)\. (.*)$`)

// TasksFile reads a change's tasks.md. Task numbering is not
// configurable, so this does not depend on p's configuration.
func (p *Parser) TasksFile(path string) ([]model.Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []model.Task
	line := 0
	inFence := false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line++
		t := sc.Text()

		// Fence-aware, the same convention internal/check/check.go and
		// internal/parse/spec.go use: a task-shaped line inside a fenced
		// illustrative example must not be read as a real task.
		if strings.HasPrefix(strings.TrimSpace(t), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		m := _taskRe.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		// The regex only captures \d+, so Atoi fails only on overflow; num
		// then stays 0, which the sequencing checks in check.TaskFindings
		// report as out of sequence.
		num, _ := strconv.Atoi(m[2])
		out = append(out, model.Task{
			Num:  num,
			Text: m[3],
			Done: m[1] == "x",
			Line: line,
		})
	}
	return out, sc.Err()
}
