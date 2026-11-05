#!/usr/bin/env bash
set -euo pipefail
fail(){ echo "::error::$1"; exit 1; }
warn(){ echo "::warning::$1"; }
[ -s feature.go ] || fail "feature.go missing"
for sym in "func FeatureWrap" "hToggle" "text/event-stream"; do
  grep -q "$sym" feature.go || fail "feature.go missing $sym"
done
[ -s web/features.html ] || fail "web/features.html missing"
[ -s web/admin/features.html ] || fail "web/admin/features.html missing"
grep -q "api/features/stream" web/features.html || fail "features.html missing SSE"
if [ -f main.go ] && ! grep -q "FeatureWrap" main.go; then
  warn "main.go not wrapped; init() DefaultServeMux fallback only"
fi
echo "OK: nebula-tools NT-1 validated"
