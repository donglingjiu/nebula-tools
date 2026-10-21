#!/usr/bin/env bash
# Nova v2.2.1 —— 修复脚本：加固 CI 构建步骤
cd "$(dirname "$0")"
[ -f ci.yml.bak ] || cp .github/workflows/ci.yml ci.yml.bak 2>/dev/null

echo "═══ setup-6：重写 ci.yml（加固 Build） ═══"

cat > .github/workflows/ci.yml <<'CI_YML'
name: CI
on:
  push:
    branches: [ main ]
  workflow_dispatch:
    inputs:
      version:
        description: '发布版本号，如 v2.2.1（仅字母数字._-，不要空格）'
        required: false
        default: ''
permissions:
  contents: write
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: 生成 go.sum（本机无 Go，由 CI 补齐）
        run: go mod tidy
      - name: 静态检查
        run: go vet ./...
      - name: 单元测试 + 全链路 E2E
        run: go test -v ./...
  build:
    needs: test
    if: ${{ github.event_name == 'workflow_dispatch' }}
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        include:
          - { goos: windows, goarch: amd64, ext: .exe }
          - { goos: linux,  goarch: amd64, ext: '' }
          - { goos: darwin, goarch: arm64, ext: '' }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go mod tidy
      - run: go vet ./... && go test ./...
      - name: Build
        env:
          VER_IN: ${{ github.event.inputs.version }}
        run: |
          set -x
          VER="$VER_IN"
          # 防护1：版本号白名单，含空格/特殊字符/为空一律回退 dev
          if ! printf '%s' "$VER" | grep -qE '^[A-Za-z0-9._-]+$'; then VER="dev"; fi
          echo "构建版本: $VER"
          mkdir -p dist
          # 防护2：CGO 全关（顺带产出免 glibc 依赖的静态二进制）
          # 防护3：-X 值单引号包裹，内部有空格也不会撕裂参数
          # 防护4：set -x 回显完整命令，再出问题日志里能看到真实调用
          CGO_ENABLED=0 GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} \
            go build -trimpath -ldflags "-s -w -X 'main.Version=${VER}'" \
            -o "dist/nova-server-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.ext }}" .
      - uses: actions/upload-artifact@v4
        with:
          name: nova-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/
CI_YML

# ---------- 自检 ----------
echo "── 自检 ──"
grep -q 'fail-fast: false' .github/workflows/ci.yml && echo "  ✓ 三平台独立运行，互不连坐" || { echo "  ✗ fail-fast 缺失"; exit 1; }
grep -q 'CGO_ENABLED=0'   .github/workflows/ci.yml && echo "  ✓ CGO 全关（静态二进制）" || { echo "  ✗ CGO 缺失"; exit 1; }
grep -q "X 'main.Version" .github/workflows/ci.yml && echo "  ✓ -X 值已加引号防护" || { echo "  ✗ 引号防护缺失"; exit 1; }
grep -q 'set -x'          .github/workflows/ci.yml && echo "  ✓ set -x 命令回显" || { echo "  ✗ set -x 缺失"; exit 1; }

git add .github/workflows/ci.yml
git commit -m "fix: 加固 CI 构建步骤（版本号白名单/-X引号防护/CGO关/命令回显）" || echo "  没有变更"
read -r -p "确认推送? [Y/n] " a
case "$a" in n|N) echo "已跳过，稍后手动 git push"; exit 0;; esac
git push && echo "═══ 已推送。重新 Run workflow，version 只填 v2.2.1 ═══"
