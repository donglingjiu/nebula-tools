#!/usr/bin/env bash
# 用法: assemble.sh <goos> <goarch> <ext> <version>
# 环境变量 FULL=1 时额外组装 full 包（随包 MySQL/Python；下载失败自动降级为 core）
set -euo pipefail
GOOS=$1; GOARCH=$2; EXT=$3; VER=$4
FULL=${FULL:-1}
cd "$(dirname "$0")/.."
source scripts/components.env
CACHE=release/cache; mkdir -p "$CACHE"
BIN="release/bin/nebula-$GOOS-$GOARCH$EXT"
[ -f "$BIN" ] || { echo "✖ 二进制不存在: $BIN"; exit 1; }

fetch_mysql() {
  local d="$CACHE/mysql-$GOOS-$GOARCH"
  [ -d "$d" ] && { echo "$d"; return 0; }
  mkdir -p "$d"
  local U F
  case "$GOOS" in
    linux)   U="https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-$MYSQL_VERSION-linux-glibc2.28-x86_64.tar.xz"; F=xz ;;
    darwin)  U="https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-$MYSQL_VERSION-macos13-arm64.tar.gz";        F=gz ;;
    windows) U="https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-$MYSQL_VERSION-winx64.zip";                  F=zip ;;
  esac
  curl -fL --retry 2 --max-time 300 "$U" -o "$CACHE/mysql.bin" || { rm -rf "$d"; return 1; }
  case $F in
    zip) mkdir -p "$CACHE/mtmp/root"; unzip -qo "$CACHE/mysql.bin" -d "$CACHE/mtmp"
         cp -r "$CACHE"/mtmp/mysql-*/. "$CACHE/mtmp/root/"; cp -r "$CACHE/mtmp/root/." "$d/"; rm -rf "$CACHE/mtmp" ;;
    xz)  tar -xJf "$CACHE/mysql.bin" -C "$d" --strip-components=1 ;;
    gz)  tar -xzf "$CACHE/mysql.bin" -C "$d" --strip-components=1 ;;
  esac
  rm -f "$CACHE/mysql.bin"; echo "$d"
}

fetch_python() {
  local d="$CACHE/python-$GOOS-$GOARCH"
  [ -d "$d" ] && { echo "$d"; return 0; }
  mkdir -p "$d"
  local U
  case "$GOOS" in
    linux)  U="https://github.com/indygreg/python-build-standalone/releases/download/20240415/cpython-$PYTHON_VERSION+20240415-x86_64-unknown-linux-gnu-install_only.tar.gz" ;;
    darwin) U="https://github.com/indygreg/python-build-standalone/releases/download/20240415/cpython-$PYTHON_VERSION+20240415-aarch64-apple-darwin-install_only.tar.gz" ;;
    windows)U="https://www.python.org/ftp/python/$PYTHON_VERSION/python-$PYTHON_VERSION-embed-amd64.zip" ;;
  esac
  curl -fL --retry 2 --max-time 300 "$U" -o "$CACHE/py.bin" || { rm -rf "$d"; return 1; }
  case "$GOOS" in
    windows) unzip -qo "$CACHE/py.bin" -d "$d" ;;
    *)       tar -xzf "$CACHE/py.bin" -C "$d" --strip-components=1 ;;
  esac
  rm -f "$CACHE/py.bin"; echo "$d"
}

pack() { # $1=后缀 $2=是否含 runtime
  local SUF=$1 RT=$2
  local NAME="nebula-portable-$VER-$GOOS-$GOARCH$SUF"
  local DIR="release/$NAME"
  rm -rf "$DIR"; mkdir -p "$DIR/config" "$DIR/data" "$DIR/sql/init"
  cp "$BIN" "$DIR/nebula$EXT"
  echo "$VER" > "$DIR/VERSION"
  printf 'driver: sqlite\n' > "$DIR/config/db.yaml"
  [ -f sql/init/01-nebula.sql ] && cp sql/init/01-nebula.sql "$DIR/sql/init/"
  if [ "$RT" = 1 ]; then
    mkdir -p "$DIR/runtime"
    if d=$(fetch_mysql); then cp -r "$d" "$DIR/runtime/mysql"; echo "      runtime/mysql ✓"; else echo "      ⚠ MySQL 下载失败，本包降级"; fi
    if d=$(fetch_python); then cp -r "$d" "$DIR/runtime/python"; echo "      runtime/python ✓"; else echo "      ⚠ Python 下载失败，本包降级"; fi
  fi
  if [ "$GOOS" = windows ]; then
    cp scripts/win/install.bat scripts/win/start.bat scripts/win/stop.bat "$DIR/"
    (cd release && zip -qry "$NAME.zip" "$NAME")
    echo "  ✔ $NAME.zip"
  else
    cp scripts/install.sh scripts/stop.sh "$DIR/"
    chmod +x "$DIR/install.sh" "$DIR/stop.sh" "$DIR/nebula"
    tar czf "release/$NAME.tar.gz" -C release "$NAME"
    echo "  ✔ $NAME.tar.gz"
  fi
  rm -rf "$DIR"
}

echo "→ 组装 core 包（SQLite 零依赖）"
pack "" 0
if [ "$FULL" = 1 ]; then
  echo "→ 组装 full 包（随包 MySQL/Python）"
  pack "-full" 1
fi
echo "→ $GOOS/$GOARCH 完成"
