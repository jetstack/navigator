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

GO_IMPORTS=$(goimports -d \
    -local github.com/jetstack/navigator \
    ${LINT_PKGS} \
)

if [ -n "${GO_IMPORTS}" ] ; then \
    echo "Please run ./hack/update-lint.sh"; \
    echo "$GO_IMPORTS"; \
    exit 1; \
fi
