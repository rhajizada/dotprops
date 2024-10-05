export PATH := $(HOME)/go/bin:$(PATH)

.PHONY: all
## all: Default target - runs fmt, vet, lint, staticcheck, and test.
all: fmt vet lint staticcheck test

.PHONY: fmt
## fmt: Formats the code using go fmt.
fmt:
	@go fmt ./...

.PHONY: vet
## vet: Examines the code for potential issues using go vet.
vet:
	@go vet ./...

.PHONY: lint
## lint: Checks the code for style mistakes with golint.
lint:
	@if ! [ -x "$$(which golint)" ]; then \
		echo "golint not found, installing..."; \
		go install golang.org/x/lint/golint@latest; \
	fi
	@golint ./...

.PHONY: staticcheck
## staticcheck: Performs advanced static analysis using staticcheck.
staticcheck:
	@if ! [ -x "$$(which staticcheck)" ]; then \
		echo "staticcheck not found, installing..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@staticcheck ./...

.PHONY: test
## test: Runs all tests using go test.
test:
	@go test ./...

.PHONY: help
## help: Show help message
help: Makefile
	@echo
	@echo " Available targets:"
	@echo
	@sed -n 's/^## //p' $< | column -t -s ':' | sed -e 's/^/  /'
