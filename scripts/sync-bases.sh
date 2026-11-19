#!/usr/bin/env bash
# OPT-029: refresh embedded base contracts from local clones (~/fanguohui/bases/)
set -euo pipefail
BSRC="${BASES:-$HOME/fanguohui/bases}"
D="$(cd "$(dirname "$0")/.." && pwd)/internal/skills"
cp "$BSRC/web-delivery-team/pipeline.json" "$BSRC/web-delivery-team/meta-info.json" "$BSRC/web-delivery-team/stage-gates.yaml" "$D/web-delivery-team/"
cp "$BSRC/product-expert-team/routing.json" "$BSRC/product-expert-team/meta-info.json" "$D/product-expert-team/"
cp "$BSRC/frontend-builder/contract.json" "$BSRC/frontend-builder/meta-info.json" "$D/frontend-builder/"
cp "$BSRC/AutoDev/contract.json" "$BSRC/AutoDev/meta-info.json" "$D/AutoDev/"
echo "[OK] contracts synced: $BSRC -> $D"
