#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

dep ensure -no-vendor -dry-run
