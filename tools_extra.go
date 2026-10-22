package main

// P0-1 微工具带（纯标准库）：经 extraRunners 挂入 runTool，tools.go 仅加一处 fallback 钩子

import (
"crypto/hmac"
"crypto/sha1"
"crypto/sha256"
"crypto/sha512"
"encoding/base32"
"encoding/base64"
"encoding/hex"
"encoding/json"
"fmt"
"hash"
"strconv"
"strings"
"time"
"unicode"
)

var extraRunners = map[string]func(string, map[string]string) (string, error){}

func init() {
reg := func(id, name, icon, desc string) {
ToolRegistry[id] = Tool{id, name, icon, "dev", desc, "text"}
}
reg("dev.hmac", "SHA-512 / HMAC", "shield", "SHA-512 摘要；填 key 则为 HMAC（alg=sha1/sha256/sha512）")
reg("dev.totp", "TOTP 动态口令", "clock", "RFC6238 生成/校验（Base32 密钥，±1 窗口）")
reg("dev.jwt", "JWT 解码", "lock", "Header/Payload 解析与过期提醒（不验签）")
reg("dev.case", "命名风格转换", "terminal", "mode=camel/pascal/snake/kebab/upper/title")
reg("dev.slug", "Slug 生成", "copy", "URL 友好化（非 ASCII 剔除）")
reg("dev.lorem", "Lorem 假文", "fileText", "英文假文（paras=段数，默认 3）")
reg("dev.count", "字数统计", "code", "字符/单词/行/字节/中文计数")

extraRunners["dev.hmac"] = toolHMAC
extraRunners["dev.totp"] = toolTOTP
extraRunners["dev.jwt"] = toolJWT
extraRunners["dev.case"] = toolCaseX
extraRunners["dev.slug"] = toolSlug
extraRunners["dev.lorem"] = toolLorem
extraRunners["dev.count"] = toolCount
}

func toolHMAC(in string, opts map[string]string) (string, error) {
alg := strings.ToLower(strings.TrimSpace(opts["alg"]))
if alg == "" {
alg = "sha256"
}
var mk func() hash.Hash
switch alg {
case "sha1":
mk = sha1.New
case "sha256":
mk = sha256.New
case "sha512":
mk = sha512.New
default:
return "", fmt.Errorf("alg 仅支持 sha1/sha256/sha512")
}
if k := opts["key"]; k != "" {
h := hmac.New(mk, []byte(k))
h.Write([]byte(in))
return "HMAC-" + strings.ToUpper(alg) + " = " + hex.EncodeToString(h.Sum(nil)), nil
}
h := mk()
h.Write([]byte(in))
return strings.ToUpper(alg) + " = " + hex.EncodeToString(h.Sum(nil)), nil
}

func xtotp6(key []byte, counter int64) string {
var buf [8]byte
c := uint64(counter)
for i := 7; i >= 0; i-- {
buf[i] = byte(c)
c >>= 8
}
m := hmac.New(sha1.New, key)
m.Write(buf[:])
s := m.Sum(nil)
o := s[len(s)-1] & 0x0f
v := (int(s[o])&0x7f)<<24 | int(s[o+1])<<16 | int(s[o+2])<<8 | int(s[o+3])
return fmt.Sprintf("%06d", v%1000000)
}

func toolTOTP(in string, opts map[string]string) (string, error) {
secret := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(in), " ", ""))
if secret == "" {
return "", fmt.Errorf("input 填 Base32 密钥")
}
key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
if err != nil {
key, err = base32.StdEncoding.DecodeString(secret)
}
if err != nil {
return "", fmt.Errorf("Base32 密钥无效: %v", err)
}
now := time.Now().Unix()
if code := strings.TrimSpace(opts["code"]); code != "" {
for _, d := range []int64{-1, 0, 1} {
if xtotp6(key, (now+d)/30) == code {
return fmt.Sprintf("校验通过 ✓（当前码 %s，窗口 ±1）", xtotp6(key, now/30)), nil
}
}
return "", fmt.Errorf("校验失败，当前正确码 %s", xtotp6(key, now/30))
}
return fmt.Sprintf("当前验证码: %s\n剩余 %d 秒\n密钥: %s", xtotp6(key, now/30), 30-now%30, secret), nil
}

func xpretty(m map[string]any) string {
if m == nil {
return "（非 JSON）"
}
b, err := json.MarshalIndent(m, "", "  ")
if err != nil {
return "（解析失败）"
}
return string(b)
}

func toolJWT(in string, opts map[string]string) (string, error) {
_ = opts
parts := strings.Split(strings.TrimSpace(in), ".")
if len(parts) < 2 {
return "", fmt.Errorf("JWT 需至少 header.payload 两段")
}
dec := func(s string) (string, error) {
if b, e := base64.RawURLEncoding.DecodeString(s); e == nil {
return string(b), nil
}
if b, e := base64.URLEncoding.DecodeString(s); e == nil {
return string(b), nil
}
return "", fmt.Errorf("base64url 解码失败")
}
h, err := dec(parts[0])
if err != nil {
return "", fmt.Errorf("header: %v", err)
}
p, err := dec(parts[1])
if err != nil {
return "", fmt.Errorf("payload: %v", err)
}
var hp, pp map[string]any
_ = json.Unmarshal([]byte(h), &hp)
_ = json.Unmarshal([]byte(p), &pp)
out := "Header:\n" + xpretty(hp) + "\n\nPayload:\n" + xpretty(pp)
if e, ok := pp["exp"].(float64); ok {
t := time.Unix(int64(e), 0)
if time.Now().After(t) {
out += "\n\n⚠ 已过期: " + t.Format("2006-01-02 15:04:05")
} else {
out += "\n\n有效期至: " + t.Format("2006-01-02 15:04:05")
}
}
return out + "\n\n（签名未校验：解码工具不持有密钥）", nil
}

func xwords(s string) []string {
var out []string
var cur strings.Builder
prev := rune(0)
flush := func() {
if cur.Len() > 0 {
out = append(out, cur.String())
cur.Reset()
}
}
for _, r := range s {
isW := unicode.IsLetter(r) || unicode.IsDigit(r)
if !isW {
flush()
prev = 0
continue
}
if prev != 0 && unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
flush()
}
cur.WriteRune(r)
prev = r
}
flush()
return out
}

func xcap(s string) string {
r := []rune(strings.ToLower(s))
if len(r) == 0 {
return ""
}
r[0] = unicode.ToUpper(r[0])
return string(r)
}

func toolCaseX(in string, opts map[string]string) (string, error) {
mode := strings.ToLower(strings.TrimSpace(opts["mode"]))
if mode == "" {
mode = "snake"
}
ws := xwords(in)
if len(ws) == 0 {
return "", fmt.Errorf("输入无有效单词")
}
lo := make([]string, len(ws))
for i, w := range ws {
lo[i] = strings.ToLower(w)
}
switch mode {
case "camel":
out := []string{lo[0]}
for _, w := range lo[1:] {
out = append(out, xcap(w))
}
return strings.Join(out, ""), nil
case "pascal":
out := make([]string, len(lo))
for i, w := range lo {
out[i] = xcap(w)
}
return strings.Join(out, ""), nil
case "snake":
return strings.Join(lo, "_"), nil
case "kebab":
return strings.Join(lo, "-"), nil
case "upper":
return strings.ToUpper(strings.Join(lo, "_")), nil
case "title":
out := make([]string, len(lo))
for i, w := range lo {
out[i] = xcap(w)
}
return strings.Join(out, " "), nil
}
return "", fmt.Errorf("mode 可选 camel/pascal/snake/kebab/upper/title")
}

func toolSlug(in string, opts map[string]string) (string, error) {
_ = opts
var b strings.Builder
lastDash := true
for _, r := range strings.ToLower(strings.TrimSpace(in)) {
if (unicode.IsLetter(r) && r < 128) || unicode.IsDigit(r) {
b.WriteRune(r)
lastDash = false
} else if !lastDash {
b.WriteByte('-')
lastDash = true
}
}
s := strings.Trim(b.String(), "-")
if s == "" {
return "", fmt.Errorf("输入不含 ASCII 字母/数字")
}
return s, nil
}

var loremSent = []string{
"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.",
"Nisi ut aliquip ex ea commodo consequat duis aute irure dolor.",
"In reprehenderit in voluptate velit esse cillum dolore eu fugiat.",
"Nulla pariatur excepteur sint occaecat cupidatat non proident.",
"Sunt in culpa qui officia deserunt mollit anim id est laborum.",
"Et harum quidem rerum facilis est et expedita distinctio nam libero.",
}

func toolLorem(in string, opts map[string]string) (string, error) {
_ = in
n, err := strconv.Atoi(strings.TrimSpace(opts["paras"]))
if err != nil || n < 1 {
n = 3
}
if n > 50 {
n = 50
}
var sb strings.Builder
for p := 0; p < n; p++ {
for i := 0; i < 4; i++ {
sb.WriteString(loremSent[(p*3+i*2)%len(loremSent)])
sb.WriteByte(' ')
}
sb.WriteString("\n\n")
}
return strings.TrimSpace(sb.String()), nil
}

func toolCount(in string, opts map[string]string) (string, error) {
_ = opts
if in == "" {
return "", fmt.Errorf("input 为空")
}
runes, cjk, words := 0, 0, 0
lines := strings.Count(in, "\n") + 1
prevSpace := true
for _, r := range in {
runes++
if unicode.Is(unicode.Han, r) {
cjk++
}
if unicode.IsLetter(r) || unicode.IsDigit(r) {
if prevSpace {
words++
}
prevSpace = false
} else {
prevSpace = true
}
}
return fmt.Sprintf("字符数: %d\n单词数: %d\n行数: %d\n字节数: %d\n中文字符: %d", runes, words, lines, len(in), cjk), nil
}
