PROJECT_NAME := $(shell grep 'module ' go.mod | awk -F"/" '{print $$NF}')
NAME := $(notdir $(PROJECT_NAME))
NEW_NAME :=$(shell echo $(NAME) | tr '_' '-')
NOW := $(shell date +'%Y%m%d%H%M%S')
TAG := $(shell git describe --always --tags --abbrev=0 | tr -d "[\r\n]")
COMMIT := $(shell git rev-parse --short HEAD| tr -d "[ \r\n\']")
VERSION_PKG := main
LD_FLAGS_BASE := -X $(VERSION_PKG).serviceName=$(NAME) -X $(VERSION_PKG).version=$(TAG)-$(COMMIT) -X $(VERSION_PKG).buildTime=$(shell date +%Y%m%d-%H%M%S) -X main.version=$(TAG)-$(COMMIT) -X 'main.buildTime=$(shell date +%Y-%m-%d\ %H:%M:%S)'
LD_FLAGS := -s -w $(LD_FLAGS_BASE)

IMPORTANT_GO_ENV_VARS := "GOPATH|GO111MODULE|GOARCH|GOCACHE|GOMODCACHE|GONOPROXY|GONOSUMDB|GOPRIVATE|GOPROXY|GOSUMDB|GOMOD|CGO"

.PHONY: all
all: binary merge-tool

.PHONY: binary
binary: export CGO_ENABLED=1
binary:
	@echo "\n###### building $(NAME)"
	@go env | grep -E $(IMPORTANT_GO_ENV_VARS)
	go build -o $(NAME) -ldflags="$(LD_FLAGS)"

.PHONY: merge-tool
merge-tool:
	go build -o merge-tool -ldflags="$(LD_FLAGS)" ./cmd/merge-tool

.PHONY: debug
debug:
	@echo "\n###### building debug binary $(NAME)"
	@go env | grep -E $(IMPORTANT_GO_ENV_VARS)
	@go version
	go build -gcflags "all=-N -l" -ldflags="$(LD_FLAGS_BASE)" -o $(NAME)

clean:
	@rm -f $(NAME)
	@rm -f $(NAME).tar.gz
	@rm -f merge-tool

fmt:
	command -v gofumpt || (WORK=$(shell pwd) && cd /tmp && GO111MODULE=on go get mvdan.cc/gofumpt && cd $(WORK))
	gofumpt -w -s -d .
	goimports -w -d .

lint:
	golangci-lint run  -v
