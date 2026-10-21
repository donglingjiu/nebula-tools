package main

import (
"crypto/aes"
"crypto/cipher"
"crypto/rand"
"encoding/base64"
"io"
"os"
"sync"
)

// At-rest secret protection (β). Key lives in data/.secret (0600, 32B
// random, generated on first use): stealing one data file alone yields
// nothing; stealing data/ wholesale equals having plaintext — the honest
// threat model for a self-hosted single binary.

var (
secOnce sync.Once
secKey  []byte
)

func secretKey() []byte {
secOnce.Do(func() {
if b, err := os.ReadFile("data/.secret"); err == nil && len(b) == 32 {
secKey = b
return
}
b := make([]byte, 32)
_, _ = io.ReadFull(rand.Reader, b)
_ = os.MkdirAll("data", 0o755)
_ = os.WriteFile("data/.secret", b, 0o600)
secKey = b
})
return secKey
}

func newGCM() (cipher.AEAD, error) {
blk, err := aes.NewCipher(secretKey())
if err != nil {
return nil, err
}
return cipher.NewGCM(blk)
}

// encSecret → "enc1:" + base64(nonce|GCM ct). decSecret tolerates legacy
// plaintext (no prefix → unchanged), so old files keep working and get
// upgraded to ciphertext on their next save.
func encSecret(plain string) string {
if plain == "" {
return ""
}
gcm, err := newGCM()
if err != nil {
return plain
}
nonce := make([]byte, gcm.NonceSize())
_, _ = io.ReadFull(rand.Reader, nonce)
out := gcm.Seal(nonce, nonce, []byte(plain), nil)
return "enc1:" + base64.StdEncoding.EncodeToString(out)
}

func decSecret(s string) string {
if len(s) < 6 || s[:5] != "enc1:" {
return s
}
raw, err := base64.StdEncoding.DecodeString(s[5:])
if err != nil {
return ""
}
gcm, err := newGCM()
if err != nil {
return ""
}
ns := gcm.NonceSize()
if len(raw) < ns {
return ""
}
pt, err := gcm.Open(nil, raw[:ns], raw[ns:], nil)
if err != nil {
return ""
}
return string(pt)
}

// maskSecret for outbound display: keep only the tail 4 chars.
func maskSecret(k string) string {
if k == "" {
return ""
}
if len(k) <= 4 {
return "****"
}
return "****" + k[len(k)-4:]
}

// llmDisk mirrors llmStore's on-disk shape without copying its mutex
// (go vet copylocks would fail `go test` on a struct copy).
type llmDisk struct {
Profiles []LLMProfile `json:"profiles"`
}
