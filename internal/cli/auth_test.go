package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptKey_NonTTY(t *testing.T) {
	t.Parallel()
	stdin := strings.NewReader("test_abcdefghijklmnop\n")
	var stderr bytes.Buffer
	got, err := promptKey(stdin, &stderr, "default")
	if err != nil {
		t.Fatalf("promptKey: %v", err)
	}
	if got != "test_abcdefghijklmnop" {
		t.Fatalf("got %q, want trimmed key", got)
	}
	if !strings.Contains(stderr.String(), "default") {
		t.Fatalf("stderr should reference profile name, got %q", stderr.String())
	}
}

func TestPromptKey_TrimsWhitespace(t *testing.T) {
	t.Parallel()
	stdin := strings.NewReader("  live_xxxxxxxxxxxx  \n")
	got, err := promptKey(stdin, new(bytes.Buffer), "prod")
	if err != nil {
		t.Fatalf("promptKey: %v", err)
	}
	if got != "live_xxxxxxxxxxxx" {
		t.Fatalf("got %q, want trimmed key", got)
	}
}

func TestPromptKey_EmptyStdin(t *testing.T) {
	t.Parallel()
	_, err := promptKey(strings.NewReader(""), new(bytes.Buffer), "default")
	if err == nil {
		t.Fatal("expected error on empty stdin, got nil")
	}
}

func TestValidKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"test_abcdefghijklmnopqrstuvwxyz0123456", true},
		{"live_abcdefghijklmnop", true},
		{"live_pub_abcdefghijklmnop", true},
		{"test_pub_abcdefghijklmnop", true},
		{"test_short", false},
		{"sk_test_abcdefghijk", false},
		{"sk_live_abcdefghijk", false},
		{"bogus_xxxxxxxxxxxxx", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := validKey(tc.in); got != tc.want {
			t.Errorf("validKey(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
