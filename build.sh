#!/bin/bash
# 盒上一键构建：./build.sh v2.7.1  （纯 Go 静态编译，免 gcc）
VER="${1:-dev}"
cd "$(dirname "$0")"
export PATH="$HOME/toolchain/go/bin:$PATH" GOPROXY=https://goproxy.cn,direct GOMAXPROCS=8
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X 'main.Version=${VER}'" -o /home/work/fanguohui/nova-box . && echo "BUILD-OK ${VER}"
