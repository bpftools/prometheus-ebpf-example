.PHONY: build push

DOCKER ?= docker
IMAGE_NAME ?= docker.io/bpftools/prometheus-ebpf-example

COMMIT_NO := $(shell git rev-parse HEAD 2> /dev/null || true)
GIT_COMMIT := $(if $(shell git status --porcelain --untracked-files=no),${COMMIT_NO}-dirty,${COMMIT_NO})
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_BRANCH_CLEAN := $(shell echo $(GIT_BRANCH) | sed -e "s/[^[:alnum:]]/-/g")

IMAGE_COMMIT := $(IMAGE_NAME):$(GIT_COMMIT)
IMAGE_BRANCH := $(IMAGE_NAME):$(GIT_BRANCH_CLEAN)
IMAGE_LATEST := $(IMAGE_NAME):latest

build:
	$(DOCKER) build -t $(IMAGE_COMMIT) .
	$(DOCKER) tag $(IMAGE_COMMIT) $(IMAGE_BRANCH)
	$(DOCKER) tag $(IMAGE_COMMIT) $(IMAGE_LATEST)

push:
	$(DOCKER) push $(IMAGE_COMMIT)
	$(DOCKER) push $(IMAGE_BRANCH)
	$(DOCKER) push $(IMAGE_LATEST)
