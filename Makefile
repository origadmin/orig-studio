# Makefile for the orig-cms project.

# ---------------------------------------------------------------------------- #
#                             Project Configuration                            #
# ---------------------------------------------------------------------------- #

# List of applications to build/release, corresponding to directories in ./cmd/
APPS := svc-api-gateway svc-user svc-media svc-content

# Path to the third_party directory for external protobufs
THIRD_PARTY_PATH := third_party

# The import path for the shared version package within your framework.
# This is used by the linker to inject build information.
VERSION_PACKAGE_PATH := github.com/origadmin/toolkits/version

# ---------------------------------------------------------------------------- #
#                         Tooling & Initialization                         #
# ---------------------------------------------------------------------------- #

.PHONY: init
init: ## 🔧 Install all required development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/google/wire/cmd/wire@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install entgo.io/ent/cmd/ent@latest
	@go install github.com/envoyproxy/protoc-gen-validate@latest
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	@echo "Running go mod tidy to ensure tool dependencies are in go.mod..."
	@go mod tidy


# ---------------------------------------------------------------------------- #
#                       Development Lifecycle Targets                      #
# ---------------------------------------------------------------------------- #

.PHONY: all deps generate clean

all: deps generate test lint ## ✅ Run all essential development steps

deps: ## 📦 Update and export third-party protobuf dependencies
	@echo "Updating buf dependencies..."
	@buf dep update
	@echo "Exporting protobuf dependencies to $(THIRD_PARTY_PATH)..."
	@buf export buf.build/bufbuild/protovalidate -o $(THIRD_PARTY_PATH)
	@buf export buf.build/protocolbuffers/wellknowntypes -o $(THIRD_PARTY_PATH)
	@buf export buf.build/googleapis/googleapis -o $(THIRD_PARTY_PATH)
	# Add other third-party dependencies as needed, e.g., kratos/apis, origadmin/runtime


generate: ## 🧬 Run all code generation tasks (Protobuf, Wire, Ent)
	@echo "Generating API Protobuf code using buf..."
	@buf generate
	@echo "Generating Common Config Protobuf code..."
	@protoc -I. -I./$(THIRD_PARTY_PATH) --go_out=paths=source_relative:. --validate_out=paths=source_relative,lang=go:. configs/*.proto
	@echo "Generating Wire code for dependency injection..."
	@go generate ./cmd/...
	@echo "Running go generate for Ent schemas..."
	@go generate ./internal/svc-user/data/...
	# Add go generate for other services' ent schemas here
	@echo "Running go mod tidy after code generation..."
	@go mod tidy

clean: ## 🧹 Clean up build artifacts and temporary files
	@echo "Cleaning up..."
	@rm -rf ./bin ./dist ./coverage.out
	@rm -rf ./api/proto/v1/*.pb.go ./api/proto/v1/*.pb.validate.go
	@rm -rf ./configs/*.pb.go ./configs/*.pb.validate.go
	@rm -rf ./internal/svc-user/data/ent
	@rm -rf ./third_party


# ---------------------------------------------------------------------------- #
#                             Testing & Linting                              #
# ---------------------------------------------------------------------------- #

.PHONY: test lint

test: ## 🧪 Run all Go tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

lint: ## 🧹 Lint the codebase with golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...


# ---------------------------------------------------------------------------- #
#                           Build & Release Variables                          #
# ---------------------------------------------------------------------------- #

GOHOSTOS ?= $(shell go env GOHOSTOS)

# Common Git information
GIT_COMMIT      := $(shell git rev-parse --short HEAD)
GIT_BRANCH      := $(shell git rev-parse --abbrev-ref HEAD)
GIT_VERSION     := $(shell git describe --tags --always)
# Get the tag at the current commit. It might be empty.
GIT_HEAD_TAG    := $(shell git tag --points-at HEAD 2>/dev/null)

# OS-specific variables for build date, tree state, and the final version tag
ifeq ($(GOHOSTOS), windows)
    BUILD_DATE   := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'")
    GIT_TREE_STATE := $(shell powershell -Command "if ((git status --porcelain)) { 'dirty' } else { 'clean' }")
    # Use the tag if it exists, otherwise use the short commit hash.
    GIT_TAG      := $(shell powershell -Command "if ('${GIT_HEAD_TAG}') { '${GIT_HEAD_TAG}' } else { '${GIT_COMMIT}' }")
else
    BUILD_DATE   := $(shell TZ=Asia/Shanghai date +%FT%T%z)
    # Check for uncommitted changes. git status --porcelain is reliable.
    GIT_TREE_STATE := $(if $(shell git status --porcelain),dirty,clean)
    # Use the tag if it exists, otherwise use the short commit hash.
    GIT_TAG      := $(if $(GIT_HEAD_TAG),$(GIT_HEAD_TAG),$(GIT_COMMIT))
endif

# If the tree is dirty, append a suffix to the version string.
ifneq ($(GIT_TREE_STATE), clean)
    GIT_VERSION := $(GIT_VERSION)-dirty
endif

# Linker flags to inject version information into the binary
LDFLAGS := -X '$(VERSION_PACKAGE_PATH).Version=$(GIT_VERSION)' \
           -X '$(VERSION_PACKAGE_PATH).GitTag=$(GIT_TAG)' \
           -X '$(VERSION_PACKAGE_PATH).GitCommit=$(GIT_COMMIT)' \
           -X '$(VERSION_PACKAGE_PATH).GitBranch=$(GIT_BRANCH)' \
           -X '$(VERSION_PACKAGE_PATH).GitTreeState=$(GIT_TREE_STATE)' \
           -X '$(VERSION_PACKAGE_PATH).BuildDate=$(BUILD_DATE)'


# ---------------------------------------------------------------------------- #
#                  Version & Build Targets (LDFLAGS Usage Demo)                  #
# ---------------------------------------------------------------------------- #

.PHONY: version build release build-docker build-docker-all
.PHONY: $(addprefix build-, $(APPS)) $(addprefix release-, $(APPS)) $(addprefix build-docker-, $(APPS))

version: ## ℹ️ Compile and run a version demo to show injected variables
	@echo "Compiling and running version demo..."
	@go run -ldflags="$(LDFLAGS)" ./cmd/version # Assuming a 'version' cmd exists for demo

build: $(addprefix build-, $(APPS)) ## 🔨 Build all application binaries with version info
release: $(addprefix release-, $(APPS)) ## 🚀 Create new releases for all applications
build-docker: $(addprefix build-docker-, $(APPS)) ## 🐳 Build all Docker images for applications

# Pattern rule for building a single application binary
# This injects the LDFLAGS into the final binary.
build-%:
	@echo "--> Building application binary: $*"
	@go build -ldflags="$(LDFLAGS)" -o ./bin/$* ./cmd/$*

# Pattern rule for building a single Docker image
build-docker-%:
	@echo "--> Building Docker image for: $*"
	@docker build . -t ghcr.io/origadmin/orig-cms/$*:latest --build-arg service_name=$*

# Pattern rule for releasing a single application via GoReleaser
release-%:
	@echo "--> Releasing application: $*"
	@goreleaser release --clean --config ./.goreleaser.yml --id $*


# ---------------------------------------------------------------------------- #
#                                     Help                                     #
# ---------------------------------------------------------------------------- #

.PHONY: help
help: ## ✨ Show this help message
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Common Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  \033[36m%-22s\033[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
