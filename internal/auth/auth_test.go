package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/99designs/keyring"
)

func TestStoreGetFileKeyring(t *testing.T) {
	t.Parallel()

	// Fixtures use synthetic data and a public test password. The oversized
	// fixture expands beyond 250 KiB; the iteration fixture requests 1300001
	// PBKDF2 rounds, just above the default limit, to keep regressions bounded.
	for _, tc := range []struct {
		name    string
		wantErr string
	}{
		{name: "legacy"},
		{name: "compressed"},
		{name: "oversized", wantErr: "Deflate stream size exceeded limit"},
		{name: "excessive-iterations", wantErr: "expected 'p2c' to be less than"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := os.ReadFile(filepath.Join("testdata", tc.name+".jwe"))
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "default"), []byte(strings.TrimSpace(string(token))), 0o600); err != nil {
				t.Fatal(err)
			}
			ring, err := keyring.Open(keyring.Config{
				AllowedBackends: []keyring.BackendType{keyring.FileBackend},
				FileDir:         dir,
				FilePasswordFunc: func(string) (string, error) {
					return "synthetic-password", nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			store := &Store{ring: ring}
			got, err := store.Get("default")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("Get error = %v, want %q", err, tc.wantErr)
				}
				if got != "" {
					t.Fatal("Get returned data for a rejected keyring entry")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != "synthetic-key" {
				t.Fatalf("Get = %q, want synthetic-key", got)
			}
		})
	}
}
