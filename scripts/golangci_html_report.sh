#!/bin/bash
set -x
TMP_FILE="scripts/report.html"
golangci-lint run -c scripts/golangci.yml > ${TMP_FILE}
open ${TMP_FILE}
