# Makefile for the orig-cms project.

# ---------------------------------------------------------------------------- #
#                             Project Configuration                            #
# ---------------------------------------------------------------------------- #

# List of applications to build/release, corresponding to directories in ./cmd/
APPS := server

# Path to the third_party directory for external protobufs (optional, created by make deps)
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

deps: ## 📦 Export third-party protobuf dependencies (optional, requires network)
	@echo "Exporting buf dependencies to $(THIRD_PARTY_PATH)..."
	@mkdir -p $(THIRD_PARTY_PATH)
	@buf export buf.build/bufbuild/protovalidate -o $(THIRD_PARTY_PATH)
	@buf export buf.build/protocolbuffers/wellknowntypes -o $(THIRD_PARTY_PATH)
	@buf export buf.build/googleapis/googleapis -o $(THIRD_PARTY_PATH)


# ---------------------------------------------------------------------------- #
#                            Code Generation                                  #
# ---------------------------------------------------------------------------- #

.PHONY: generate gen-proto gen-openapi gen-types gen-convpb gen-wire gen-ent

generate: ## 🧬 Run all code generation tasks (Proto → OpenAPI → Types → Wire → Ent → ConvPB)
	@$(MAKE) gen-proto
	@$(MAKE) gen-wire
	@$(MAKE) gen-ent
	@$(MAKE) gen-convpb
	@echo "Running go mod tidy..."
	@go mod tidy

gen-proto: ## 📡 Generate Protobuf Go code + OpenAPI docs via buf (requires network for remote deps)
	@echo "Generating API Protobuf code + OpenAPI documentation..."
	@cd api && buf generate
	@echo "✅ OpenAPI docs generated at docs/api/openapi.yaml"

gen-openapi: ## 📖 Generate OpenAPI/Swagger documentation only
	@$(MAKE) gen-proto

gen-types: ## 🔷 Generate TypeScript types from OpenAPI v3 spec
	@echo "Generating TypeScript types from OpenAPI v3..."
	@if not exist "docs/api/openapi.yaml" ( \
		echo "⚠️  OpenAPI file not found. Run 'make gen-proto' first." && exit 1 \
	)
	@call scripts\gen-types.bat

gen-wire: ## 🔌 Generate Wire dependency injection code
	@echo "Generating Wire code..."
	@go generate ./cmd/...

gen-ent: ## 🏗️ Generate Ent ORM code (Schema: internal/data/entity/schema/)
	@echo "Generating Ent schemas..."
	@go generate ./internal/data/entity/...

gen-convpb: ## 🔀 Generate entity↔proto type conversion code
	@echo "Generating convpb type converters..."
	@go generate ./internal/data/convpb

clean: ## 🧹 Clean up build artifacts and generated code
	@echo "Cleaning up..."
	@$(call RMDIR,./bin)
	@$(call RMDIR,./dist)
	@$(call RMDIR,./coverage.out)
	@$(call RMDIR,./api/gen)
	@$(call RMDIR,./docs/api)
	@$(call RMDIR,./web/src/types/api.ts)
	@$(call RMDIR,./internal/data/entity/ent)
	@$(call RMDIR,./third_party)


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
    GIT_TAG      := $(shell powershell -Command "if ('${GIT_HEAD_TAG}') { '${GIT_HEAD_TAG}' } else { '${GIT_COMMIT}' }")
    RMDIR        = powershell -Command "if (Test-Path '$(1)') { Remove-Item -Recurse -Force '$(1)' }"
    CPDIR        = powershell -Command "Copy-Item -Recurse -Path '$(1)' -Destination '$(2)'"
    EXE_SUFFIX   := .exe
else
    BUILD_DATE   := $(shell TZ=Asia/Shanghai date +%FT%T%z)
    GIT_TREE_STATE := $(if $(shell git status --porcelain),dirty,clean)
    GIT_TAG      := $(if $(GIT_HEAD_TAG),$(GIT_HEAD_TAG),$(GIT_COMMIT))
    RMDIR        = rm -rf $(1)
    CPDIR        = cp -r $(1) $(2)
    EXE_SUFFIX   :=
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
#                     Frontend & Server Build Targets                         #
# ---------------------------------------------------------------------------- #

.PHONY: build-frontend build-server build-server-dev run-server run-server-dev

build-frontend: ## 🎨 Build frontend React SPA and copy to embed directory
	@echo "==> Building frontend..."
	cd web && bun install && bun run build
	@echo "==> Copying web/dist -> internal/frontend/dist/"
	@$(call RMDIR,internal/frontend/dist)
	@$(call CPDIR,web/dist,internal/frontend/dist)
	@echo "✅ Frontend build complete."

build-server: build-frontend ## 📦 Build production server (frontend embedded, single binary)
	@echo "==> Building server binary (production, frontend embedded)..."
	@go build -ldflags="$(LDFLAGS)" -o ./bin/server$(EXE_SUFFIX) ./cmd/server
	@echo ""
	@echo "✅ Build complete!"
	@echo "   Binary: ./bin/server$(EXE_SUFFIX)"
	@echo "   Run:    ./bin/server$(EXE_SUFFIX) -conf configs"
	@echo ""

build-server-dev: ## 🔧 Build dev server (frontend from disk, no embed)
	@echo "==> Building server binary (dev mode, frontend from disk)..."
	@go build -tags dev -ldflags="$(LDFLAGS)" -o ./bin/server$(EXE_SUFFIX) ./cmd/server
	@echo ""
	@echo "✅ Build complete!"
	@echo "   Binary: ./bin/server$(EXE_SUFFIX)"
	@echo "   Run:    ./bin/server$(EXE_SUFFIX) -conf configs"
	@echo ""

run-server: build-server ## 🚀 Build and run production server
	@echo "==> Starting server..."
	@./bin/server$(EXE_SUFFIX) -conf configs

run-server-dev: build-server-dev ## 🏃 Build and run dev server
	@echo "==> Starting dev server..."
	@./bin/server$(EXE_SUFFIX) -conf configs


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
