#!/bin/sh
# S1a semantic-coverage + held-out compiler audits (one command).
set -eu
cd "$(dirname "$0")/.."
exec go test ./internal/memory -count=1 -timeout 180s \
  -run 'TestSemanticCoverageAudit|TestHeldOutCompilerCoverageAudit|TestHeldOutNamedSubjectAndAddresseeAudit|TestNoBenchmarkSurfaceFormsInProductCode'
