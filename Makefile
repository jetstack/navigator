BINDIR        ?= bin
HACK_DIR     ?= hack
NAVIGATOR_PKG = github.com/jetstack-experimental/navigator

TYPES_FILES    = $(shell find pkg/apis -name types.go)

REGISTRY := jetstackexperimental
IMAGE_NAME := navigator
BUILD_TAG := build
IMAGE_TAGS := canary

BUILD_IMAGE_DIR := hack/builder
BUILD_IMAGE_NAME := navigator/builder

GOPATH ?= /tmp/go

.get_deps:
	@echo "Grabbing dependencies..."
	@go get -d k8s.io/kubernetes/cmd/libs/go2idl/... || true
	@go get -d github.com/kubernetes/repo-infra || true
	@touch $@

help:
	# all 		- runs verify, build and docker_build targets
	# test 		- runs go_test target
	# build 	- runs generate, and then go_build targets
	# generate 	- generates pkg/client/ files
	# verify 	- verifies generated files & scripts

# Util targets
##############
.PHONY: all test verify

all: verify build docker_build

test: go_test

build: generate go_build

generate: .generate_files

verify: .hack_verify go_verify

.hack_verify: .generate_exes
	@echo Running repo-infra verify scripts
	@echo Running href checker:
	@${HACK_DIR}/verify-links.sh
	@echo Running errexit checker:
	@${HACK_DIR}/verify-errexit.sh
	@echo Running generated client checker:
	@${HACK_DIR}/verify-client-gen.sh

# Builder image targets
#######################
docker_%: .builder_image
	docker run -it \
		-v ${GOPATH}/src:/go/src \
		-v $(shell pwd):/go/src/${NAVIGATOR_PKG} \
		-w /go/src/${NAVIGATOR_PKG} \
		-e GOPATH=/go \
		${BUILD_IMAGE_NAME} \
		/bin/sh -c "make $*"

.builder_image:
	docker build -t ${BUILD_IMAGE_NAME} ${BUILD_IMAGE_DIR}

# Docker targets
################
docker_build:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) .

docker_push: docker_build
	set -e; \
		for tag in $(IMAGE_TAGS); do \
		docker tag $(REGISTRY)/$(IMAGE_NAME):$(BUILD_TAG) $(REGISTRY)/$(IMAGE_NAME):$${tag} ; \
		docker push $(REGISTRY)/$(IMAGE_NAME):$${tag}; \
	done


# Go targets
#################
go_verify: go_fmt go_vet go_test

go_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o navigator_linux_amd64 ./cmd/controller

go_test:
	go test $$(go list ./... | grep -v '/vendor/')

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

go_vet:
	go vet $$(go list ./... | grep -v '/vendor/')

# This section contains the code generation stuff
#################################################
.generate_exes: .get_deps \
				$(BINDIR)/defaulter-gen \
                $(BINDIR)/deepcopy-gen \
                $(BINDIR)/conversion-gen \
                $(BINDIR)/client-gen \
                $(BINDIR)/lister-gen \
                $(BINDIR)/informer-gen
	touch $@

$(BINDIR)/defaulter-gen:
	go build -o $@ k8s.io/kubernetes/cmd/libs/go2idl/defaulter-gen

$(BINDIR)/deepcopy-gen:
	go build -o $@ k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen

$(BINDIR)/conversion-gen:
	go build -o $@ k8s.io/kubernetes/cmd/libs/go2idl/conversion-gen

$(BINDIR)/client-gen:
	go build -o $@ k8s.io/kubernetes/cmd/libs/go2idl/client-gen

$(BINDIR)/lister-gen:
	go build -o $@ k8s.io/kubernetes/cmd/libs/go2idl/lister-gen

$(BINDIR)/informer-gen:
	go build -o $@ k8s.io/kubernetes/cmd/libs/go2idl/informer-gen

# Regenerate all files if the gen exes changed or any "types.go" files changed
.generate_files: .generate_exes $(TYPES_FILES)
	# Generate defaults
	$(BINDIR)/defaulter-gen \
		--v 1 --logtostderr \
		--go-header-file "$${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal" \
		--input-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal/v1alpha1" \
		--extra-peer-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal" \
		--extra-peer-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal/v1alpha1" \
		--output-file-base "zz_generated.defaults"
	# Generate deep copies
	$(BINDIR)/deepcopy-gen \
		--v 1 --logtostderr \
		--go-header-file "$${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal" \
		--input-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal/v1alpha1" \
		--bounding-dirs "github.com/openshift/open-service-broker-sdk" \
		--output-file-base zz_generated.deepcopy
	# Generate conversions
	$(BINDIR)/conversion-gen \
		--v 1 --logtostderr \
		--go-header-file "$${GOPATH}/src/github.com/kubernetes/repo-infra/verify/boilerplate/boilerplate.go.txt" \
		--input-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal" \
		--input-dirs "$(NAVIGATOR_PKG)/pkg/apis/marshal/v1alpha1" \
		--output-file-base zz_generated.conversion
	# generate all pkg/client contents
	$(HACK_DIR)/update-client-gen.sh
