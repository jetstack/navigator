SHELL := /bin/bash
BINDIR        ?= bin
HACK_DIR     ?= hack
NAVIGATOR_PKG = github.com/jetstack/navigator

TYPES_FILES      = $(shell find pkg/apis -name types.go)

REGISTRY := jetstackexperimental
IMAGE_NAME := navigator
BUILD_TAG := build
IMAGE_TAGS := canary
CHART_VALUES := ${HACK_DIR}/testdata/values.yaml

BUILD_IMAGE_DIR := hack/builder
BUILD_IMAGE_NAME := navigator/builder

CMDS := controller apiserver pilot-elasticsearch

GOPATH ?= /tmp/go

help:
	# all       - runs verify, build and docker_build targets
	# test      - runs go_test target
	# e2e-test  - runs e2e tests
	# build     - runs generate, and then go_build targets
	# generate  - generates pkg/client/ files
	# verify    - verifies generated files & scripts

# Util targets
##############
.PHONY: all test verify $(CMDS) generate

all: verify build docker_build

test: go_test

.run_e2e:
	export CHART_VALUES=${CHART_VALUES}; \
	${HACK_DIR}/prepare-e2e.sh; \
	${HACK_DIR}/e2e.sh

.e2e_init:
	${HACK_DIR}/install-e2e-dependencies.sh

e2e-test: .e2e_init build docker_build .run_e2e

build: $(CMDS)

generate: .generate_files

verify: .hack_verify go_verify

.hack_verify:
	@echo Running repo-infra verify scripts
	@echo Running href checker:
	@${HACK_DIR}/verify-links.sh
	@echo Running errexit checker:
	@${HACK_DIR}/verify-errexit.sh
	@echo Running generated client checker:
	@${HACK_DIR}/verify-client-gen.sh

# Docker targets
################
DOCKER_BUILD_TARGETS = $(addprefix docker_build_, $(CMDS))
$(DOCKER_BUILD_TARGETS):
	$(eval DOCKER_BUILD_CMD := $(subst docker_build_,,$@))
	eval $$(minikube docker-env --profile $$HOSTNAME --shell sh); \
	docker build -t $(REGISTRY)/$(IMAGE_NAME)-$(DOCKER_BUILD_CMD):$(BUILD_TAG) -f Dockerfile.$(DOCKER_BUILD_CMD) .
docker_build: $(DOCKER_BUILD_TARGETS)

DOCKER_PUSH_TARGETS = $(addprefix docker_push_, $(CMDS))
$(DOCKER_PUSH_TARGETS):
	$(eval DOCKER_PUSH_CMD := $(subst docker_push_,,$@))
	set -e; \
		for tag in $(IMAGE_TAGS); do \
		eval $$(minikube docker-env --profile $$HOSTNAME --shell sh); \
		docker tag $(REGISTRY)/$(IMAGE_NAME)-$(DOCKER_PUSH_CMD):$(BUILD_TAG) $(REGISTRY)/$(IMAGE_NAME)-$(DOCKER_PUSH_CMD):$${tag} ; \
		docker push $(REGISTRY)/$(IMAGE_NAME)-$(DOCKER_PUSH_CMD):$${tag}; \
	done
docker_push: $(DOCKER_PUSH_TARGETS)

# Go targets
#################
go_verify: go_fmt go_test go_build

$(CMDS):
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o navigator-$@_linux_amd64 ./cmd/$@

go_build: $(CMDS)

go_test:
	go test -v $$(go list ./... | grep -v '/vendor/')

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

# This section contains the code generation stuff
#################################################
# Regenerate all files if any "types.go" files changed
.generate_files: $(TYPES_FILES)
	# generate all pkg/client contents
	$(HACK_DIR)/update-client-gen.sh
