GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

LD_FLAGS ?= "-extldflags '-static'"

DOCKER_IMAGE ?= roobre/prosody-httpupload
DOCKER_TAG ?= latest

BIN := prosody-httpupload-$(GOOS)-$(GOARCH)

build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-buildmode=exe -ldflags $(LD_FLAGS) \
		-o $(BIN) \
		./cmd

.PHONY: image
image: build
	DOCKER_BUILDKIT=1 docker build . -t $(DOCKER_IMAGE):$(DOCKER_TAG)
