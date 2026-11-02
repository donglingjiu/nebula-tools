#!/bin/bash
# 数据备份: tar打包数据目录, 时间戳命名, 保留最近N份
# 用法: bash scripts/backup.sh [数据目录] [备份目录] [保留份数]
# 默认: data/ -> backups/ 保留14份
# 恢复: 停服 -> tar -xzf backups/nova-data-<ts>.tar.gz -C 仓库根 -> 覆盖data/ -> 重启
# 注: SQLite wal/shm随目录一并打包, 恢复后自动recover; 建议低峰或停服时备份
set -euo pipefail
SRC="${1:-data}"
DST="${2:-backups}"
KEEP="${3:-14}"
[ -d "$SRC" ] || { echo "数据目录不存在: $SRC" >&2; exit 1; }
mkdir -p "$DST"
ts=$(date +%Y%m%d-%H%M%S)
out="$DST/nova-data-$ts.tar.gz"
tar -czf "$out" -C "$(dirname "$SRC")" "$(basename "$SRC")"
echo "备份完成: $out ($(du -h "$out" | cut -f1))"
ls -1t "$DST"/nova-data-*.tar.gz | tail -n +$((KEEP+1)) | xargs -r rm --
echo "当前保留: $(ls -1 "$DST"/nova-data-*.tar.gz | wc -l) 份 (上限 $KEEP)"
