#!/bin/bash

# This script builds the MCMS related contracts and copy the compiled binaries to the
# destination directory. The destination directory is used by the CTF e2e tests to
# deploy the programs on Solana.

# Usage: ./e2e/tests/solana/compile-mcm-contracts.sh

set -euo pipefail

REPO_URL="https://github.com/smartcontractkit/chainlink-ccip"
REPO_DIR="chainlink-ccip"

PROJECT_ROOT=$(git rev-parse --show-toplevel)
if [[ -z "${PROJECT_ROOT}" ]]; then
  echo "Error: This script must be run within a Git repository."
  exit 1
fi

GO_MOD_FILE="${PROJECT_ROOT}/go.mod"
if [[ ! -f "$GO_MOD_FILE" ]]; then
  echo "Error: go.mod file not found in the current directory."
  exit 1
fi

PROGRAM_DIR="chains/solana/contracts/target/deploy"
DEST_DIR="${PROJECT_ROOT}/tests/integration/solana/artifacts"
TEMP_DIR=$(mktemp -d)

# Parse the go.mod file for the specific entry
MOD_ENTRY=$(grep -E 'github\.com/smartcontractkit/chainlink-ccip/chains/solana\s+v[0-9]+\.[0-9]+\.[0-9]+-[0-9]+-[a-f0-9]+' "$GO_MOD_FILE")
if [[ -z "$MOD_ENTRY" ]]; then
  echo "Error: Could not find the required entry in go.mod."
  exit 1
fi

# Extract repo URL and pseudo-version
PSEUDO_VERSION=$(echo "$MOD_ENTRY" | awk '{print $2}')

# Extract commit SHA from pseudo-version (last 12 characters)
COMMIT_HASH=$(echo "$PSEUDO_VERSION" | grep -oE '[a-f0-9]{12}$')
if [[ -z "$COMMIT_HASH" ]]; then
  echo "Error: Could not extract commit SHA from pseudo-version: $PSEUDO_VERSION"
  exit 1
fi

# Programs to build
PROGRAMS=("timelock" "access-controller" "external-program-cpi-stub")

cd "${PROJECT_ROOT}"

git clone "${REPO_URL}" "${TEMP_DIR}/${REPO_DIR}"
cd "${TEMP_DIR}/${REPO_DIR}"
git checkout "${COMMIT_HASH}"

for program in "${PROGRAMS[@]}"; do
  cd "chains/solana/contracts/programs/${program}"
  cargo build-sbf
  cd -
done

mkdir -p "${DEST_DIR}"
cp -r "${TEMP_DIR}/${REPO_DIR}/${PROGRAM_DIR}/"* "${DEST_DIR}/"

rm -rf "${TEMP_DIR}"

echo "MCMS contracts compiled successfully"