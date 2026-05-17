// Package auth resolves the Lob API key for a given invocation. Precedence
// (highest first): explicit --api-key flag, LOB_API_KEY env, keyring entry
// for the active profile.
package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/voska/loby/internal/errfmt"
)

const (
	serviceName    = "loby"
	envVar         = "LOB_API_KEY"
	keyringPassEnv = "LOBY_KEYRING_PASSWORD"
)

// Source describes where a resolved key came from. Useful for `auth status`.
type Source string

// Possible Source values surfaced by Resolve.
const (
	SourceFlag    Source = "flag"
	SourceEnv     Source = "env"
	SourceKeyring Source = "keyring"
	SourceUnset   Source = "unset"
)

// Resolved is the outcome of Resolve — the active key with provenance.
type Resolved struct {
	Key     string
	Profile string
	Source  Source
}

// ErrNotConfigured is returned by Resolve when no key is available.
var ErrNotConfigured = errors.New("no Lob API key configured: run `loby auth login`")

// Store provides keyring-backed persistence. Each profile is one keyring item.
type Store struct {
	ring keyring.Keyring
}

// Open opens (or creates) the loby keyring. The released binaries are
// CGO-free, so the platform-native backends (macOS Keychain, secret-service,
// WinCred) aren't compiled in — every platform lands on the encrypted file
// backend under $XDG_CONFIG_HOME/loby, unlocked via filePassword().
func Open() (*Store, error) {
	cfgDir, err := loobyConfigDir()
	if err != nil {
		return nil, err
	}
	ring, err := keyring.Open(keyring.Config{
		ServiceName:              serviceName,
		FileDir:                  cfgDir,
		FilePasswordFunc:         filePassword,
		KeychainTrustApplication: true,
		KeychainSynchronizable:   false,
		LibSecretCollectionName:  "default",
		WinCredPrefix:            serviceName,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.SecretServiceBackend,
			keyring.WinCredBackend,
			keyring.FileBackend,
		},
	})
	if err != nil {
		return nil, errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("open keyring: %w", err))
	}
	return &Store{ring: ring}, nil
}

// Set stores apiKey under profile.
func (s *Store) Set(profile, apiKey string) error {
	if profile == "" {
		profile = "default"
	}
	err := s.ring.Set(keyring.Item{
		Key:         profile,
		Data:        []byte(apiKey),
		Label:       "loby (" + profile + ")",
		Description: "Lob API key",
	})
	if err != nil {
		return errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("keyring set: %w", err))
	}
	return nil
}

// Get retrieves the key for profile. Returns ErrNotConfigured if absent.
func (s *Store) Get(profile string) (string, error) {
	if profile == "" {
		profile = "default"
	}
	item, err := s.ring.Get(profile)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", ErrNotConfigured
		}
		return "", errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("keyring get: %w", err))
	}
	return string(item.Data), nil
}

// Remove deletes profile from the keyring. Returns ErrNotConfigured if absent.
func (s *Store) Remove(profile string) error {
	if profile == "" {
		profile = "default"
	}
	if err := s.ring.Remove(profile); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return ErrNotConfigured
		}
		return errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("keyring remove: %w", err))
	}
	return nil
}

// List returns the profile names known to the keyring.
func (s *Store) List() ([]string, error) {
	names, err := s.ring.Keys()
	if err != nil {
		return nil, errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("keyring keys: %w", err))
	}
	return names, nil
}

// Resolve returns the active key for this invocation, honoring the precedence
// flag > env > keyring. profileName defaults to "default". A nil store means
// keyring lookup is skipped (useful for tests).
func Resolve(flag, profileName string, store *Store) (Resolved, error) {
	if flag != "" {
		return Resolved{Key: flag, Profile: profileName, Source: SourceFlag}, nil
	}
	if v := os.Getenv(envVar); v != "" {
		return Resolved{Key: v, Profile: profileName, Source: SourceEnv}, nil
	}
	if store == nil {
		return Resolved{Profile: profileName, Source: SourceUnset}, ErrNotConfigured
	}
	key, err := store.Get(profileName)
	if err != nil {
		return Resolved{Profile: profileName, Source: SourceUnset}, err
	}
	return Resolved{Key: key, Profile: profileName, Source: SourceKeyring}, nil
}

// Environment classifies an API key as test or live by its prefix.
// Lob uses `live_…` / `test_…` for secret keys and `live_pub_…` /
// `test_pub_…` for publishable; the live/test classification is the
// same in both cases.
func Environment(key string) string {
	switch {
	case strings.HasPrefix(key, "live_"):
		return "live"
	case strings.HasPrefix(key, "test_"):
		return "test"
	case key == "":
		return ""
	default:
		return "unknown"
	}
}

// Prefix returns a safely-truncated, non-secret prefix of key for display.
func Prefix(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:8] + "…"
}

// filePassword unlocks the encrypted-file keyring backend. The native OS
// keychains are preferred — this is only reached when none of Keychain,
// SecretService, or WinCred are usable (headless boxes, unsigned binaries
// macOS denies, BSDs). LOBY_KEYRING_PASSWORD lets CI / scripted setups
// avoid the interactive prompt.
func filePassword(prompt string) (string, error) {
	if p := os.Getenv(keyringPassEnv); p != "" {
		return p, nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errfmt.Wrap(errfmt.ConfigError, fmt.Errorf(
			"file-backed keyring requires a password; set %s or run interactively", keyringPassEnv,
		))
	}
	fmt.Fprintf(os.Stderr, "%s: ", prompt)
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("read keyring password: %w", err))
	}
	return strings.TrimSpace(string(raw)), nil
}

func loobyConfigDir() (string, error) {
	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("home dir: %w", err))
		}
		xdg = home + "/.config"
	}
	dir := filepath.Join(xdg, "loby")
	if err := os.MkdirAll(dir, 0o700); err != nil { //nolint:gosec // dir is built from $XDG_CONFIG_HOME or os.UserHomeDir, both trusted
		return "", errfmt.Wrap(errfmt.ConfigError, fmt.Errorf("mkdir %s: %w", dir, err))
	}
	return dir, nil
}
