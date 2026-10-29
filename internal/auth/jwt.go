package auth

import (
"crypto/hmac"
"crypto/rand"
"crypto/sha256"
"crypto/subtle"
"encoding/base64"
"encoding/json"
"errors"
"os"
"path/filepath"
"strings"
"time"
)

type claims struct {
Sub   string `json:"sub"`
Email string `json:"email"`
Role  string `json:"role"`
Iat   int64  `json:"iat"`
Exp   int64  `json:"exp"`
}

var jwtSecret []byte

// initSecret：NB_JWT_SECRET 优先；否则 data/secret.key 持久化随机密钥（重启不掉 to
func initSecret(dir string) error {
if s := os.Getenv("NB_JWT_SECRET"); s
p := filepath.Join(dir, "secret.key")
if b, err := os.ReadFile(p); err == nil && len(b) >= 32 { jwtSecret = b; return nil }
b := make([]byte, 32)
if _, e
if err := os.WriteFile(p, b, 0o600); err != nil { return err }
jwtSecret = b
return nil
}

func b64url(b []by

var errBadToken = erro

func signToken(u *User, ttl time.Duration) (string, error) {
now := time.Now()
body, err := json.Marshal(claims{Sub: u.ID, Email: u.Email, Role: u.Role, Iat: now.Unix(), Exp: now.Add(ttl).Unix()})
if err != nil { return "", err }
payload := b64url([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + b64url(body)
mac := hmac.New(sha256.New, jwtSecret)
mac.Write([]byte(payload))
return payload + "." + b64url(mac.Sum(nil)), nil
}

func parseToken(tok string) (*claims, error) {
p := strings.Split(tok, ".")
if len(p) != 3 { return nil, errBadToken }
mac := hmac.New(sha256.New, jwtSecret)
mac.Write([]byte(p[0] + "." + p[1]))
if subtle.ConstantTimeCompare([]byte(b64url(mac.Sum(nil))), []byte(p[2])) != 1 { return nil, errBadToken }
raw, err := base64.RawURLEncoding.DecodeString(p[1])
if err != nil { return nil, errBadToken }
var c claims
if json.Unmarshal(raw, &c) != nil || c.Exp < time.Now().Unix() { return nil, errBadToken }
return &c, nil
}
