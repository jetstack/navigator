#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE}")/..
cd "${REPO_ROOT}"

# TODO: remove nasty grepping
LINT_PKGS=$(find . -type f \
    ! -name 'zz_generated.*' \
    ! -path './pkg/client/*' \
    ! -path './vendor/*' | grep '\.go'
)

goimports -w \
    -local github.com/jetstack/navigator \
    ${LINT_PKGS}
