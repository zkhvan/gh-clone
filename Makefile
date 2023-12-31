MAIN_PACKAGE_PATH := .
BINARY_NAME := gh-clone

# ============================================================================
# HELPERS
# ============================================================================

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ============================================================================
# QUALITY CONTROL
# ============================================================================

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## lint: lint the code
.PHONY: lint
lint:
	golangci-lint run

## lint/fix: lint the code, auto-fix if possible
.PHONY: lint/fix
lint/fix:
	golangci-lint run --fix

# ============================================================================
# DEVELOPMENT
# ============================================================================

## test: run all tests
.PHONY: test
test:
	go test -v -race ./...

## build: build the application
.PHONY: build
build:
	go build -o=${BINARY_NAME} ${MAIN_PACKAGE_PATH}
