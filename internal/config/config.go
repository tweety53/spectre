// Package config reads spectre/config.md, the one file that changes how a
// tree behaves. Every key is closed and every value validated: a typo that
// silently disabled a rule is the failure this file would otherwise add.
package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RuleNames is the closed set of validation rules, in file order.
var RuleNames = []string{
	"headings",
	"placeholders",
	"shall-clause",
	"malformed-bullet",
	"id-sequence",
	"task-sequence",
	"refs",
}

// Config is one tree's settings. The zero value is not valid — its rules
// map is nil, so RuleOn reports every rule off; build one with Default or
// Parse instead.
type Config struct {
	rules      map[string]bool // rule name -> produces findings; read via RuleOn, changed via WithRule
	Modal      string
	IDPrefix   string
	SpecsDir   string
	ChangesDir string
	Extension  string
}

// RuleOn reports whether name is currently enabled.
func (c Config) RuleOn(name string) bool {
	return c.rules[name]
}

// WithRule returns a copy of c with name's rule set to on. The copy's rule
// map is cloned, never shared with c's, so mutating it — the only way to
// change a rule — can never alias back into c or any other copy taken from
// the same tree. Config's map field is unexported for exactly this reason:
// "c := t.Cfg; c.rules[name] = v" would otherwise silently mutate t.Cfg's
// own map, since a Go map copies by reference.
func (c Config) WithRule(name string, on bool) Config {
	c.rules = maps.Clone(c.rules)
	c.rules[name] = on
	return c
}

// Default is the configuration of a tree with no config.md.
func Default() Config {
	rules := make(map[string]bool, len(RuleNames))
	for _, name := range RuleNames {
		rules[name] = true
	}
	return Config{
		rules:      rules,
		Modal:      "SHALL",
		IDPrefix:   "R",
		SpecsDir:   "specs",
		ChangesDir: "changes",
		Extension:  ".md",
	}
}

var (
	_bulletRe   = regexp.MustCompile(`^- ([a-z][a-z-]*): (.+)$`)
	_idPrefixRe = regexp.MustCompile(`^[A-Za-z][A-Za-z-]*$`)
)

// Load reads <root>/config.md. An absent file is the default configuration.
func Load(root string) (Config, error) {
	raw, err := os.ReadFile(filepath.Join(root, "config.md"))
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}
	return Parse(raw)
}

// Parse reads a config.md body. Every error names the offending line.
// Fenced code blocks are skipped, the same convention internal/check uses,
// so an illustrative example inside a ``` fence never changes a real
// setting. A key set twice in the same section is an error naming both
// lines, the same shape tree.Peers uses for a duplicate peer name.
func Parse(raw []byte) (Config, error) {
	c := Default()
	section := ""
	inFence := false
	seen := map[string]int{} // "section|key" -> line first set
	for i, line := range strings.Split(string(raw), "\n") {
		at := fmt.Sprintf("config.md:%d:", i+1)
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "# "):
			continue
		case strings.HasPrefix(trimmed, "## "):
			section = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			if section != "Rules" && section != "Vocabulary" && section != "Layout" {
				return Config{}, fmt.Errorf("%s unknown section %q, want one of \"Rules\", \"Vocabulary\", \"Layout\"", at, section)
			}
			continue
		case !strings.HasPrefix(trimmed, "- "):
			continue
		}

		m := _bulletRe.FindStringSubmatch(trimmed)
		if m == nil {
			return Config{}, fmt.Errorf("%s malformed setting %q, want \"- <key>: <value>\"", at, trimmed)
		}
		key, value := m[1], strings.TrimSpace(m[2])

		seenKey := section + "|" + key
		if prevLine, dup := seen[seenKey]; dup {
			noun := "key"
			if section == "Rules" {
				noun = "rule"
			}
			return Config{}, fmt.Errorf("%s duplicate %s %q (first seen on line %d)", at, noun, key, prevLine)
		}
		seen[seenKey] = i + 1

		switch section {
		case "Rules":
			if _, known := c.rules[key]; !known {
				return Config{}, fmt.Errorf("%s unknown rule %q, want one of %s", at, key, strings.Join(RuleNames, ", "))
			}
			switch value {
			case "error":
				c.rules[key] = true
			case "off":
				c.rules[key] = false
			default:
				return Config{}, fmt.Errorf("%s rule %q: want \"error\" or \"off\", got %q", at, key, value)
			}
		case "Vocabulary":
			switch key {
			case "modal":
				if strings.ContainsAny(value, " \t") {
					return Config{}, fmt.Errorf("%s modal must be a single word, got %q", at, value)
				}
				c.Modal = value
			case "id-prefix":
				if !_idPrefixRe.MatchString(value) {
					return Config{}, fmt.Errorf("%s id-prefix must start with a letter and hold only letters and hyphens, got %q", at, value)
				}
				c.IDPrefix = value
			default:
				return Config{}, fmt.Errorf("%s unknown key %q in Vocabulary, want \"modal\" or \"id-prefix\"", at, key)
			}
		case "Layout":
			switch key {
			case "specs", "changes":
				if err := checkRelative(value); err != nil {
					return Config{}, fmt.Errorf("%s %s: %v", at, key, err)
				}
				if key == "specs" {
					c.SpecsDir = value
				} else {
					c.ChangesDir = value
				}
			case "extension":
				if !strings.HasPrefix(value, ".") {
					return Config{}, fmt.Errorf("%s extension must start with a dot, got %q", at, value)
				}
				c.Extension = value
			default:
				return Config{}, fmt.Errorf("%s unknown key %q in Layout, want \"specs\", \"changes\" or \"extension\"", at, key)
			}
		default:
			return Config{}, fmt.Errorf("%s setting outside any section", at)
		}
	}
	return c, nil
}

func checkRelative(p string) error {
	if p == "" {
		return errors.New("must not be empty")
	}
	if filepath.IsAbs(p) {
		return fmt.Errorf("must be relative to the tree, got %q", p)
	}
	clean := filepath.Clean(p)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("must stay inside the tree, got %q", p)
	}
	return nil
}
