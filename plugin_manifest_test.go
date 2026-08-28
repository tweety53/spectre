package spectre

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// boundaryAfter reports whether token appears in text followed by a
// character that could not continue a directory name or invocation
// string, or by the end of the text — so a token that is a prefix of a
// longer one (e.g. "spectre" in "spectrecorp") does not falsely match.
func boundaryAfter(text, token string) bool {
	pattern := regexp.QuoteMeta(token) + `([^A-Za-z0-9_.-]|$)`
	return regexp.MustCompile(pattern).MatchString(text)
}

// cpDestinationLine returns the line in readme that copies into
// ~/.claude/skills — selected by what it is for, not by being the
// first "cp -r" line in document order, so an unrelated "cp -r"
// example elsewhere in the README cannot be mistaken for the
// manual-install command. It fails the test unless there is exactly
// one such line.
func cpDestinationLine(t *testing.T, readme string) string {
	t.Helper()
	var lines []string
	for _, line := range strings.Split(readme, "\n") {
		if strings.Contains(line, "cp -r") && strings.Contains(line, "~/.claude/skills") {
			lines = append(lines, line)
		}
	}
	if len(lines) != 1 {
		t.Fatalf("README.md has %d lines that \"cp -r\" into ~/.claude/skills, want exactly 1", len(lines))
	}
	return lines[0]
}

// repoRoot resolves the repository root from this test file's own
// location: `go test` can be invoked from anywhere, and this file sits
// at the root it is resolving, so there is no ".." to walk.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine this test file's own path")
	}
	return filepath.Dir(thisFile)
}

type pluginManifest struct {
	Name   string `json:"name"`
	Skills string `json:"skills"`
}

type marketplaceManifest struct {
	Name    string `json:"name"`
	Plugins []struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	} `json:"plugins"`
}

func TestPluginManifestsAgree(t *testing.T) {
	root := repoRoot(t)
	pluginPath := filepath.Join(root, ".claude-plugin", "plugin.json")
	marketPath := filepath.Join(root, ".claude-plugin", "marketplace.json")

	var plugin pluginManifest
	var market marketplaceManifest

	t.Run("manifests parse", func(t *testing.T) {
		for path, into := range map[string]any{pluginPath: &plugin, marketPath: &market} {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			if err := json.Unmarshal(raw, into); err != nil {
				t.Fatalf("%s is not valid JSON: %v", path, err)
			}
		}
	})

	t.Run("plugin name matches", func(t *testing.T) {
		if len(market.Plugins) != 1 {
			t.Fatalf("%s lists %d plugins, want exactly 1", marketPath, len(market.Plugins))
		}
		if market.Plugins[0].Name != plugin.Name {
			t.Errorf("marketplace names plugin %q, plugin.json names itself %q",
				market.Plugins[0].Name, plugin.Name)
		}
	})

	t.Run("marketplace source resolves", func(t *testing.T) {
		if len(market.Plugins) == 0 {
			t.Fatalf("%s lists no plugins", marketPath)
		}
		src := filepath.Join(root, market.Plugins[0].Source, ".claude-plugin", "plugin.json")
		srcInfo, err := os.Stat(src)
		if err != nil {
			t.Fatalf("marketplace source %q does not hold a plugin manifest: %v",
				market.Plugins[0].Source, err)
		}
		pluginInfo, err := os.Stat(pluginPath)
		if err != nil {
			t.Fatalf("stat %s: %v", pluginPath, err)
		}
		if !os.SameFile(srcInfo, pluginInfo) {
			t.Errorf("marketplace source %q resolves to a plugin manifest other than this repository's own %s",
				market.Plugins[0].Source, pluginPath)
		}
	})

	t.Run("skills path names the real skill directory", func(t *testing.T) {
		resolved, err := filepath.Abs(filepath.Join(root, plugin.Skills))
		if err != nil {
			t.Fatalf("resolving plugin.json skills %q: %v", plugin.Skills, err)
		}
		want, err := filepath.Abs(filepath.Join(root, ".claude", "skills"))
		if err != nil {
			t.Fatalf("resolving the repository's own .claude/skills: %v", err)
		}
		if resolved != want {
			t.Errorf("plugin.json skills %q resolves to %q, want the repository's own %q",
				plugin.Skills, resolved, want)
		}
	})

	t.Run("README install command matches", func(t *testing.T) {
		readme, err := os.ReadFile(filepath.Join(root, "README.md"))
		if err != nil {
			t.Fatalf("reading README.md: %v", err)
		}
		want := "/plugin install " + plugin.Name + "@" + market.Name
		if !boundaryAfter(string(readme), want) {
			t.Errorf("README.md does not carry %q, which is what the manifests imply", want)
		}
	})

	t.Run("README creates the destination before copying", func(t *testing.T) {
		readme, err := os.ReadFile(filepath.Join(root, "README.md"))
		if err != nil {
			t.Fatalf("reading README.md: %v", err)
		}
		s := string(readme)
		mkdirIdx := strings.Index(s, "mkdir -p ~/.claude/skills")
		if mkdirIdx < 0 {
			t.Fatal("README.md does not contain \"mkdir -p ~/.claude/skills\"")
		}
		cpIdx := strings.Index(s, cpDestinationLine(t, s))
		if mkdirIdx > cpIdx {
			t.Errorf("README.md's \"mkdir -p ~/.claude/skills\" line (offset %d) must come before its \"cp -r\" line (offset %d)", mkdirIdx, cpIdx)
		}
	})

	t.Run("README and skill agree on the directory name", func(t *testing.T) {
		skillsDir := filepath.Join(root, ".claude", "skills")
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			t.Fatalf("reading %s: %v", skillsDir, err)
		}
		var real []string
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(skillsDir, e.Name(), "SKILL.md")); err == nil {
				real = append(real, e.Name())
			}
		}
		if len(real) == 0 {
			t.Fatalf("%s holds no <dir>/SKILL.md", skillsDir)
		}

		readme, err := os.ReadFile(filepath.Join(root, "README.md"))
		if err != nil {
			t.Fatalf("reading README.md: %v", err)
		}

		cpLine := cpDestinationLine(t, string(readme))

		for _, name := range real {
			cpTarget := filepath.Join(".claude", "skills", name)
			if !boundaryAfter(cpLine, cpTarget) {
				t.Errorf("README.md's manual-install command (%q) does not name the real skill directory %q (want it to reference %q)",
					cpLine, name, cpTarget)
			}

			own, err := os.ReadFile(filepath.Join(skillsDir, name, "SKILL.md"))
			if err != nil {
				t.Fatalf("reading %s/%s/SKILL.md: %v", skillsDir, name, err)
			}
			invoke := "/" + plugin.Name + ":" + name
			if !boundaryAfter(string(own), invoke) {
				t.Errorf("%s/SKILL.md does not claim the plugin-install invocation %q for its own directory name",
					name, invoke)
			}
		}
	})
}
