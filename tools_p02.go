package main

// P0-2 微工具带第二弹：cron/chmod/subnet/yaml，经 extraRunners 挂入 runTool

import (
"encoding/json"
"fmt"
"net"
"strconv"
"strings"
"time"

"gopkg.in/yaml.v3"
)

func init() {
reg := func(id, name, icon, desc string) {
ToolRegistry[id] = Tool{id, name, icon, "dev", desc, "text"}
}
reg("dev.cron", "Cron 计划解析", "clock", "5 段表达式：字段说明 + 未来 N 次运行（count，默认 3）")
reg("dev.chmod", "Chmod 计算", "lock", "754 ↔ rwxr-xr-- 双向互转与逐位解释")
reg("dev.subnet", "IPv4 子网计算", "globe", "CIDR → 网络/广播/掩码/主机范围")
reg("dev.yaml", "YAML ↔ JSON", "copy", "自动识别方向互转")

extraRunners["dev.cron"] = toolCron
extraRunners["dev.chmod"] = toolChmod
extraRunners["dev.subnet"] = toolSubnet
extraRunners["dev.yaml"] = toolYaml
}

/* ---------- cron ---------- */

var cronBounds = [5][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 6}}
var cronNames = [5]string{"分钟", "小时", "日", "月", "周"}

func toolCron(in string, opts map[string]string) (string, error) {
fields := strings.Fields(strings.TrimSpace(in))
if len(fields) != 5 {
return "", fmt.Errorf("需 5 段表达式（分 时 日 月 周），示例：0 9 * * 1-5")
}
var sets [5]map[int]bool
for i, f := range fields {
s, err := cronField(f, cronBounds[i][0], cronBounds[i][1])
if err != nil {
return "", fmt.Errorf("%s 字段 %q：%v", cronNames[i], f, err)
}
sets[i] = s
}
count := 3
if n, e := strconv.Atoi(strings.TrimSpace(opts["count"])); e == nil && n > 0 && n <= 20 {
count = n
}
var b strings.Builder
b.WriteString("字段解析：\n")
for i, f := range fields {
if f == "*" {
b.WriteString(fmt.Sprintf("  %s：*（任意值）\n", cronNames[i]))
} else {
b.WriteString(fmt.Sprintf("  %s：%s\n", cronNames[i], f))
}
}
b.WriteString(fmt.Sprintf("未来 %d 次运行（本地时区）：\n", count))
t := time.Now().Truncate(time.Minute)
found := 0
for step := 0; step < 527040 && found < count; step++ {
if sets[0][t.Minute()] && sets[1][t.Hour()] && sets[2][t.Day()] && sets[3][int(t.Month())] && sets[4][int(t.Weekday())] {
b.WriteString(fmt.Sprintf("  %s\n", t.Format("2006-01-02 15:04 Mon")))
found++
}
t = t.Add(time.Minute)
}
if found == 0 {
return "", fmt.Errorf("一年内无匹配运行时间，请检查表达式")
}
return strings.TrimRight(b.String(), "\n"), nil
}

func cronField(f string, lo, hi int) (map[int]bool, error) {
set := map[int]bool{}
for _, part := range strings.Split(f, ",") {
step := 1
if i := strings.Index(part, "/"); i >= 0 {
n, err := strconv.Atoi(part[i+1:])
if err != nil || n <= 0 {
return nil, fmt.Errorf("步长非法")
}
step = n
part = part[:i]
}
a, z := lo, hi
if part != "*" {
if i := strings.Index(part, "-"); i >= 0 {
x, e1 := strconv.Atoi(part[:i])
y, e2 := strconv.Atoi(part[i+1:])
if e1 != nil || e2 != nil {
return nil, fmt.Errorf("范围非法")
}
a, z = x, y
} else {
v, e := strconv.Atoi(part)
if e != nil {
return nil, fmt.Errorf("数值非法")
}
a, z = v, v
}
}
if a < lo || z > hi || a > z {
return nil, fmt.Errorf("超出范围 %d-%d", lo, hi)
}
for v := a; v <= z; v += step {
set[v] = true
}
}
return set, nil
}

/* ---------- chmod ---------- */

func toolChmod(in string, _ map[string]string) (string, error) {
s := strings.TrimSpace(in)
if len(s) == 4 && s[0] == '0' {
s = s[1:]
}
var oct, sym string
switch {
case len(s) == 3 && strings.Trim(s, "01234567") == "":
oct = s
n, _ := strconv.ParseInt(s, 8, 32)
sym = permTriplet(int(n)>>6) + permTriplet(int(n)>>3&7) + permTriplet(int(n)&7)
case len(s) == 9 && strings.Trim(s, "rwx-") == "":
sym = s
n := 0
for i := 0; i < 9; i++ {
bit := 0
switch i % 3 {
case 0:
if s[i] == 'r' {
bit = 4
}
case 1:
if s[i] == 'w' {
bit = 2
}
case 2:
if s[i] == 'x' {
bit = 1
}
}
n = n*8 + bit
}
oct = fmt.Sprintf("%o", n)
default:
return "", fmt.Errorf("需 3 位八进制（如 754）或 9 位权限串（如 rwxr-xr--）")
}
n, _ := strconv.ParseInt(oct, 8, 32)
who := [3]string{"属主 u", "属组 g", "其他 o"}
var b strings.Builder
b.WriteString(fmt.Sprintf("八进制 %s ↔ 符号 %s\n", oct, sym))
for i := 0; i < 3; i++ {
v := int(n>>(6-3*i)) & 7
b.WriteString(fmt.Sprintf("%s（%d）：%s\n", who[i], v, permDesc(v)))
}
return strings.TrimRight(b.String(), "\n"), nil
}

func permTriplet(v int) string {
t := ""
if v&4 != 0 {
t += "r"
} else {
t += "-"
}
if v&2 != 0 {
t += "w"
} else {
t += "-"
}
if v&1 != 0 {
t += "x"
} else {
t += "-"
}
return t
}

func permDesc(v int) string {
var p []string
if v&4 != 0 {
p = append(p, "读")
}
if v&2 != 0 {
p = append(p, "写")
}
if v&1 != 0 {
p = append(p, "执行")
}
if len(p) == 0 {
return "无权限"
}
return strings.Join(p, "·")
}

/* ---------- subnet ---------- */

func toolSubnet(in string, _ map[string]string) (string, error) {
s := strings.TrimSpace(in)
if s == "" {
return "", fmt.Errorf("请输入 IPv4 CIDR，示例：192.168.1.130/26")
}
ip, ipnet, err := net.ParseCIDR(s)
if err != nil {
return "", fmt.Errorf("无法解析 %q（示例 192.168.1.130/26）", s)
}
base := ip.To4()
if base == nil {
return "", fmt.Errorf("仅支持 IPv4")
}
network := base.Mask(ipnet.Mask)
ones, _ := ipnet.Mask.Size()
mask := net.CIDRMask(ones, 32)
bcast := make(net.IP, 4)
for i := 0; i < 4; i++ {
bcast[i] = network[i] | ^mask[i]
}
var hosts uint64
var first, last net.IP
switch {
case ones >= 32:
hosts, first, last = 1, network, bcast
case ones == 31:
hosts, first, last = 2, network, bcast
default:
hosts = (uint64(1) << uint(32-ones)) - 2
first = make(net.IP, 4)
copy(first, network)
first[3]++
last = make(net.IP, 4)
copy(last, bcast)
last[3]--
}
return fmt.Sprintf("网络地址：%s\n广播地址：%s\n子网掩码：%s（/%d）\n可用主机：%d 台\n主机范围：%s – %s",
network, bcast, net.IP(mask).String(), ones, hosts, first, last), nil
}

/* ---------- yaml ---------- */

func toolYaml(in string, _ map[string]string) (string, error) {
s := strings.TrimSpace(in)
if s == "" {
return "", fmt.Errorf("请输入 YAML 或 JSON 文本")
}
var v interface{}
if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
if err := json.Unmarshal([]byte(s), &v); err != nil {
return "", fmt.Errorf("JSON 解析失败：%v", err)
}
out, err := yaml.Marshal(v)
if err != nil {
return "", fmt.Errorf("YAML 生成失败：%v", err)
}
return "检测到 JSON 输入 → YAML 输出：\n" + string(out), nil
}
if err := yaml.Unmarshal([]byte(s), &v); err != nil {
return "", fmt.Errorf("YAML 解析失败：%v", err)
}
out, err := json.MarshalIndent(v, "", "  ")
if err != nil {
return "", fmt.Errorf("JSON 生成失败：%v", err)
}
return "检测到 YAML 输入 → JSON 输出：\n" + string(out), nil
}
