package cli

import (
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/voska/loby/internal/auth"
	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
)

// LobClient builds an authenticated Lob client from the active globals. Auth
// resolution follows --api-key > LOB_API_KEY > keyring(profile).
//
// Keyring open failures are surfaced unless flag/env satisfy auth on their
// own — a broken keychain should not silently masquerade as "no API key
// configured" because the recovery path is different (fix the keychain vs.
// run `auth login`).
func (g *Globals) LobClient() (*client.Client, error) {
	var (
		store    *auth.Store
		storeErr error
	)
	if g.APIKey == "" && os.Getenv("LOB_API_KEY") == "" {
		store, storeErr = auth.Open()
		if storeErr != nil {
			return nil, storeErr
		}
	}
	resolved, err := auth.Resolve(g.APIKey, g.Profile, store)
	if err != nil {
		if errors.Is(err, auth.ErrNotConfigured) {
			return nil, errfmt.Wrap(errfmt.AuthRequired, errors.New("no Lob API key configured: run `loby auth login` or set LOB_API_KEY"))
		}
		return nil, err
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if g.Debug {
		logger = slog.New(slog.NewTextHandler(g.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return client.New(resolved.Key, client.WithLogger(logger)), nil
}
