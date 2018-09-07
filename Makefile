GIT_VERSION = $(shell git describe --tags --dirty)
VERSION ?= $(GIT_VERSION)

VERSION_PACKAGE := github.com/dispatchframework/funky/pkg/version

GO_LDFLAGS := -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS += -X $(VERSION_PACKAGE).commit=$(shell git rev-parse HEAD)


.PHONY: all
all: linux

.PHONY: test
test: ## run tests
	@echo running tests...
	$(GO) test -v $(shell go list -v ./... | grep -v /vendor/ | grep -v integration )

.PHONY: linux
linux:
	GOOS=linux go build -ldflags "$(GO_LDFLAGS)" -o funky main.go

.PHONY: release
release: linux
	tar -czf funky$(VERSION).linux-amd64.tgz funky
