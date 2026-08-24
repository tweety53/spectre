package parse

import (
	"bufio"
	"os"
	"regexp"
	"strconv"

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
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line++
		m := _taskRe.FindStringSubmatch(sc.Text())
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

// TasksFile reads a change's tasks.md using the default configuration.
//
// Deprecated: build a *Parser via New and call its TasksFile method — a
// *tree.Tree carries one configured for its own tree.
func TasksFile(path string) ([]model.Task, error) { return defaultParser.TasksFile(path) }
