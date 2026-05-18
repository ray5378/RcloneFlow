#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# 1) Type-aware unused symbols
npx -y ts-prune -p tsconfig.json | tee unused.txt || true

# 2) ESLint no-unused-vars (strict)
npx -y eslint --ext .ts,.vue src --max-warnings=0 || true

echo "---"
echo "Report saved to frontend/unused.txt (ts-prune)"
