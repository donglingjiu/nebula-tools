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
[ -s pipeline.go ] || fail "pipeline.go missing"
grep -q "func PipelineWrap" pipeline.go || fail "pipeline.go missing PipelineWrap"
[ -s web/pipeline.html ] || fail "web/pipeline.html missing"
[ -s metrics.go ] || fail "metrics.go missing"
grep -q "func MetricsWrap" metrics.go || fail "metrics.go missing MetricsWrap"
[ -s web/admin-metrics.html ] || fail "web/admin-metrics.html missing"
if [ -f main.go ]; then
  grep -q "FeatureWrap" main.go || warn "main.go missing FeatureWrap"
  grep -q "PipelineWrap" main.go || warn "main.go missing PipelineWrap"
  grep -q "MetricsWrap" main.go || warn "main.go missing MetricsWrap"
fi
echo "OK: nebula-tools NT-1/3/4 validated"
