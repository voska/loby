package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// IdempotencyKey derives a deterministic Idempotency-Key for a (command, args,
// body) triple. When deterministic is false, a random suffix is appended so
// repeated invocations from a shell loop do not collide on Lob's cache.
//
// The intent: agents calling the same command with the same flags and body
// twice within 24h get the cached response from Lob. Different invocations
// produce different keys.
func IdempotencyKey(command string, flags map[string]string, body any, deterministic bool) (string, error) {
	h := sha256.New()
	h.Write([]byte(command))
	h.Write([]byte{0})

	keys := make([]string, 0, len(flags))
	for k := range flags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{'='})
		h.Write([]byte(flags[k]))
		h.Write([]byte{0})
	}

	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("idempotency: marshal body: %w", err)
		}
		h.Write(buf)
	}

	digest := hex.EncodeToString(h.Sum(nil))[:24]
	if deterministic {
		return "loby-" + digest, nil
	}

	var entropy [4]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return "", fmt.Errorf("idempotency: random: %w", err)
	}
	return "loby-" + digest + "-" + strings.ToLower(hex.EncodeToString(entropy[:])), nil
}
