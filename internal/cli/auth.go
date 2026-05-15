package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/voska/loby/internal/auth"
	"github.com/voska/loby/internal/errfmt"
)

// AuthCmd groups credential commands.
type AuthCmd struct {
	Login  AuthLoginCmd  `cmd:"" help:"Store a Lob API key in the OS keychain under a profile."`
	Logout AuthLogoutCmd `cmd:"" help:"Remove a stored API key."`
	Status AuthStatusCmd `cmd:"" help:"Print the active auth profile and key prefix."`
	List   AuthListCmd   `cmd:"" help:"List configured profiles."`
}

// AuthLoginCmd stores an API key under the active profile.
type AuthLoginCmd struct {
	Key string `help:"API key to store under the active profile. Required when --no-input is set."`
}

// Run prompts for the key (unless --key or stdin is non-TTY) and stores it.
func (c *AuthLoginCmd) Run(g *Globals) error {
	key := c.Key
	if key == "" && g.NoInput {
		return errfmt.Wrap(errfmt.UsageError, errors.New("--key is required when --no-input is set"))
	}
	if key == "" {
		k, err := promptKey(g.Stdin, g.Stderr, g.Profile)
		if err != nil {
			return err
		}
		key = k
	}
	if !validKey(key) {
		return errfmt.Wrap(errfmt.UsageError, errors.New("invalid Lob API key (must start with live_, test_, sk_live_, or sk_test_)"))
	}

	store, err := auth.Open()
	if err != nil {
		return err
	}
	if err := store.Set(g.Profile, key); err != nil {
		return err
	}
	g.Writer().Notice("stored %s key under profile %q", auth.Environment(key), g.Profile)
	return g.Writer().Render(Status{
		Profile:     g.Profile,
		Source:      string(auth.SourceKeyring),
		Configured:  true,
		KeyPrefix:   auth.Prefix(key),
		Environment: auth.Environment(key),
	})
}

// AuthLogoutCmd removes a stored API key.
type AuthLogoutCmd struct{}

// Run removes the profile's key from the keyring.
func (c *AuthLogoutCmd) Run(g *Globals) error {
	store, err := auth.Open()
	if err != nil {
		return err
	}
	if err := store.Remove(g.Profile); err != nil {
		if errors.Is(err, auth.ErrNotConfigured) {
			g.Writer().Notice("profile %q was not configured", g.Profile)
			return errfmt.Wrap(errfmt.Empty, errors.New("not configured"))
		}
		return err
	}
	g.Writer().Notice("removed profile %q", g.Profile)
	return nil
}

// AuthStatusCmd reports the active profile + key prefix without exposing secrets.
type AuthStatusCmd struct{}

// Status is the JSON shape of `loby auth status --json`.
type Status struct {
	Profile     string `json:"profile"`
	Source      string `json:"source"`
	Environment string `json:"environment,omitempty"`
	KeyPrefix   string `json:"key_prefix,omitempty"`
	Configured  bool   `json:"configured"`
}

// Run inspects the credential precedence chain and reports the result.
func (c *AuthStatusCmd) Run(g *Globals) error {
	store, _ := auth.Open()
	resolved, err := auth.Resolve(g.APIKey, g.Profile, store)
	s := Status{
		Profile:    resolved.Profile,
		Source:     string(resolved.Source),
		Configured: err == nil,
	}
	if resolved.Key != "" {
		s.KeyPrefix = auth.Prefix(resolved.Key)
		s.Environment = auth.Environment(resolved.Key)
	}
	if err != nil && !errors.Is(err, auth.ErrNotConfigured) {
		return err
	}
	return g.Writer().Render(s)
}

// AuthListCmd lists known profiles.
type AuthListCmd struct{}

// Run enumerates keyring profiles.
func (c *AuthListCmd) Run(g *Globals) error {
	store, err := auth.Open()
	if err != nil {
		return err
	}
	names, err := store.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		g.Writer().Notice("no profiles configured")
		return errfmt.Wrap(errfmt.Empty, errors.New("no profiles configured"))
	}
	type profile struct {
		Name        string `json:"name"`
		Environment string `json:"environment,omitempty"`
		KeyPrefix   string `json:"key_prefix,omitempty"`
	}
	out := make([]profile, 0, len(names))
	for _, n := range names {
		key, err := store.Get(n)
		if err != nil {
			out = append(out, profile{Name: n})
			continue
		}
		out = append(out, profile{Name: n, Environment: auth.Environment(key), KeyPrefix: auth.Prefix(key)})
	}
	return g.Writer().Render(out)
}

func promptKey(stdin io.Reader, stderr io.Writer, profile string) (string, error) {
	if stdin == nil {
		stdin = os.Stdin
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	_, _ = fmt.Fprintf(stderr, "Paste your Lob API key for profile %q (won't be echoed by your terminal if connected to a TTY):\n", profile)
	r := bufio.NewReader(stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", errfmt.Wrap(errfmt.GeneralError, fmt.Errorf("read key: %w", err))
	}
	return strings.TrimSpace(line), nil
}

func validKey(k string) bool {
	switch {
	case strings.HasPrefix(k, "live_"), strings.HasPrefix(k, "test_"),
		strings.HasPrefix(k, "sk_live_"), strings.HasPrefix(k, "sk_test_"):
		return len(k) > 10
	default:
		return false
	}
}
