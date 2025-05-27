#!/bin/bash

set -euo pipefail

REPO_URL="https://github.com/smartcontractkit/chainlink-ccip"
REPO_DIR="chainlink-ccip"

PROJECT_ROOT=$(git rev-parse --show-toplevel)
GO_MOD_FILE="${PROJECT_ROOT}/go.mod"

PROGRAM_DIR="chains/solana/contracts/target/deploy"
DEST_DIR="${PROJECT_ROOT}/tests/integration/solana/artifacts"
TEMP_DIR=$(mktemp -d)

MOD_ENTRY=$(grep -E 'github\.com/smartcontractkit/chainlink-ccip/chains/solana\s+v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+-[a-f0-9]+' "$GO_MOD_FILE")
PSEUDO_VERSION=$(echo "$MOD_ENTRY" | awk '{print $2}')
COMMIT_HASH=$(echo "$PSEUDO_VERSION" | grep -oE '[a-f0-9]{12}$')

# Programs to build
PROGRAMS=("timelock" "access-controller" "external-program-cpi-stub")

cd "${PROJECT_ROOT}"

# ✅ This was missing
git clone "${REPO_URL}" "${TEMP_DIR}/${REPO_DIR}"
cd "${TEMP_DIR}/${REPO_DIR}"
git checkout "${COMMIT_HASH}"
cd chains/solana/contracts
for program in "${PROGRAMS[@]}"; do
  LIB_NAME=$(echo "$program" | tr '-' '_')  # convert to snake_case
  solana-verify build --library-name "$LIB_NAME"
done

mkdir -p "${DEST_DIR}"
cp -r "${TEMP_DIR}/${REPO_DIR}/${PROGRAM_DIR}/"* "${DEST_DIR}/"

rm -rf "${TEMP_DIR}"

echo "Timelock programs compiled successfully"
