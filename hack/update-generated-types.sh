#!/bin/bash

GO_OUT=${GO_OUT-.}

protoc --go_out=$GO_OUT pkg/api/v1/*.proto
