package cli

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
)

// SchemaCmd implements `loby schema`. It walks the Kong AST so the output
// always matches the actual command tree — no drift between docs and binary.
type SchemaCmd struct {
	Path []string `arg:"" optional:"" help:"Optional command path (e.g. 'postcards create') to scope output."`
}

// Schema is the JSON shape returned by `loby schema --json`.
type Schema struct {
	Name        string   `json:"name"`
	Help        string   `json:"help,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Flags       []Flag   `json:"flags,omitempty"`
	Positional  []Flag   `json:"positional,omitempty"`
	Subcommands []Schema `json:"subcommands,omitempty"`
}

// Flag describes one flag or positional argument in JSON-friendly form.
type Flag struct {
	Name     string   `json:"name"`
	Short    string   `json:"short,omitempty"`
	Help     string   `json:"help,omitempty"`
	Type     string   `json:"type,omitempty"`
	Default  string   `json:"default,omitempty"`
	Env      []string `json:"env,omitempty"`
	Required bool     `json:"required,omitempty"`
	Enum     []string `json:"enum,omitempty"`
}

// Run walks the Kong model and emits a structured tree.
func (c *SchemaCmd) Run(g *Globals, k *kong.Kong) error {
	root := nodeToSchema(k.Model.Node)
	if len(c.Path) > 0 {
		scoped, ok := scopeTo(&root, c.Path)
		if !ok {
			return fmt.Errorf("unknown command: %s", strings.Join(c.Path, " "))
		}
		root = *scoped
	}
	return g.Writer().Render(root)
}

func nodeToSchema(n *kong.Node) Schema {
	s := Schema{
		Name:    n.Name,
		Help:    n.Help,
		Aliases: n.Aliases,
	}
	for _, f := range n.Flags {
		if f.Hidden {
			continue
		}
		s.Flags = append(s.Flags, flagFromKong(f))
	}
	for _, p := range n.Positional {
		s.Positional = append(s.Positional, positionalFromKong(p))
	}
	for _, sub := range n.Children {
		if sub.Hidden {
			continue
		}
		s.Subcommands = append(s.Subcommands, nodeToSchema(sub))
	}
	return s
}

func flagFromKong(f *kong.Flag) Flag {
	v := f.Value
	out := Flag{
		Name:     v.Name,
		Help:     v.Help,
		Type:     typeOf(v),
		Default:  v.Default,
		Env:      f.Envs,
		Required: v.Required,
		Enum:     enumOf(v),
	}
	if f.Short != 0 {
		out.Short = string(f.Short)
	}
	return out
}

func positionalFromKong(v *kong.Value) Flag {
	out := Flag{
		Name:     v.Name,
		Help:     v.Help,
		Type:     typeOf(v),
		Default:  v.Default,
		Required: v.Required,
		Enum:     enumOf(v),
	}
	if v.Tag != nil {
		out.Env = v.Tag.Envs
	}
	return out
}

func typeOf(v *kong.Value) string {
	if v.Tag != nil && v.Tag.Type != "" {
		return v.Tag.Type
	}
	return v.Target.Type().String()
}

func enumOf(v *kong.Value) []string {
	if v.Enum == "" {
		return nil
	}
	parts := strings.Split(v.Enum, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func scopeTo(s *Schema, path []string) (*Schema, bool) {
	if len(path) == 0 {
		return s, true
	}
	for i := range s.Subcommands {
		sub := &s.Subcommands[i]
		if sub.Name == path[0] || containsString(sub.Aliases, path[0]) {
			return scopeTo(sub, path[1:])
		}
	}
	return nil, false
}

func containsString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
