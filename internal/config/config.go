// Package config reads spectre/config.md, the one file that changes how a
// tree behaves. Every key is closed and every value validated: a typo that
// silently disabled a rule is the failure this file would otherwise add.
package config

import (
	"errors"
	"fmt"
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

// Config is one tree's settings.
type Config struct {
	Rules      map[string]bool // rule name -> produces findings
	Modal      string
	IDPrefix   string
	SpecsDir   string
	ChangesDir string
	Extension  string
}

// Default is the configuration of a tree with no config.md.
func Default() Config {
	rules := make(map[string]bool, len(RuleNames))
	for _, name := range RuleNames {
		rules[name] = true
	}
	return Config{
		Rules:      rules,
		Modal:      "SHALL",
		IDPrefix:   "R",
		SpecsDir:   "specs",
		ChangesDir: "changes",
		Extension:  ".md",
	}
}

var (
	bulletRe   = regexp.MustCompile(`^- ([a-z][a-z-]*): (.+)$`)
	idPrefixRe = regexp.MustCompile(`^[A-Za-z][A-Za-z-]*$`)
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
func Parse(raw []byte) (Config, error) {
	c := Default()
	section := ""
	for i, line := range strings.Split(string(raw), "\n") {
		at := fmt.Sprintf("config.md:%d:", i+1)
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "# "):
			continue
		case strings.HasPrefix(trimmed, "## "):
			section = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			if section != "Rules" && section != "Vocabulary" && section != "Layout" {
				return Config{}, fmt.Errorf("%s unknown section %q", at, section)
			}
			continue
		case !strings.HasPrefix(trimmed, "- "):
			continue
		}

		m := bulletRe.FindStringSubmatch(trimmed)
		if m == nil {
			return Config{}, fmt.Errorf("%s malformed setting %q, want \"- <key>: <value>\"", at, trimmed)
		}
		key, value := m[1], strings.TrimSpace(m[2])

		switch section {
		case "Rules":
			if _, known := c.Rules[key]; !known {
				return Config{}, fmt.Errorf("%s unknown rule %q", at, key)
			}
			switch value {
			case "error":
				c.Rules[key] = true
			case "off":
				c.Rules[key] = false
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
				if !idPrefixRe.MatchString(value) {
					return Config{}, fmt.Errorf("%s id-prefix must start with a letter and hold only letters and hyphens, got %q", at, value)
				}
				c.IDPrefix = value
			default:
				return Config{}, fmt.Errorf("%s unknown key %q in Vocabulary", at, key)
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
				return Config{}, fmt.Errorf("%s unknown key %q in Layout", at, key)
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
