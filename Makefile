# ==============================================================================
# VibeTerm — Production Cross-Platform Build Automation Suite
# Pure Native (CGO + OpenGL/X11/Wayland/MinGW) Build Pipeline
# ==============================================================================

APP_NAME      := vibeterm
MODULE_NAME   := vibeterm
VERSION       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
GIT_COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

DIST_DIR      := dist
BIN_DIR       := $(DIST_DIR)/bin
RELEASE_DIR   := $(DIST_DIR)/release

# Production Linker Optimization Flags (Strip symbol table, debug info & inject build metadata)
LDFLAGS_COMMON := -s -w \
	-X '$(MODULE_NAME)/internal/config.Version=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/config.GitCommit=$(GIT_COMMIT)' \
	-X '$(MODULE_NAME)/internal/config.BuildTime=$(BUILD_TIME)'

LDFLAGS_LINUX   := $(LDFLAGS_COMMON)
LDFLAGS_WINDOWS := $(LDFLAGS_COMMON) -H=windowsgui

# Toolchain definitions
CC_LINUX_AMD64   := gcc
CC_LINUX_ARM64   := aarch64-linux-gnu-gcc
CC_WINDOWS_AMD64 := x86_64-w64-mingw32-gcc
WINDRES          := x86_64-w64-mingw32-windres

.PHONY: all help deps test clean \
        build-linux-amd64 build-linux-arm64 build-windows-amd64 \
        package-linux-amd64 package-linux-arm64 package-windows-amd64 \
        release checksums docker-build docker-extract run

all: help

help:
	@echo "================================================================================"
	@echo " ⚡ VibeTerm Build System (${VERSION})"
	@echo "================================================================================"
	@echo " Available Targets:"
	@echo "   make build-linux-amd64    - Build native Linux x86_64 binary"
	@echo "   make build-linux-arm64    - Cross-compile Linux aarch64 binary (via cross-gcc)"
	@echo "   make build-windows-amd64  - Cross-compile Windows x86_64 .exe (via MinGW-w64)"
	@echo "   make package-linux-amd64  - Build & bundle Linux x86_64 release tar.gz"
	@echo "   make package-linux-arm64  - Build & bundle Linux ARM64 release tar.gz"
	@echo "   make package-windows-amd64- Build & bundle Windows x86_64 release zip"
	@echo "   make release              - Build all architectures, packages & checksums"
	@echo "   make docker-build         - Build all release artifacts deterministically in Docker"
	@echo "   make test                 - Run test suite"
	@echo "   make clean                - Wipe dist and temporary build outputs"
	@echo "================================================================================"

deps:
	go mod download
	go mod tidy

test:
	go test -v -race ./internal/terminal ./internal/config ./internal/scanner ./internal/automation ./internal/gitops ./internal/ai ./internal/forwarding

clean:
	rm -rf $(DIST_DIR)
	rm -f cmd/vibeterm/*.syso cmd/vibeterm/*.o

# ------------------------------------------------------------------------------
# 🐧 Linux Builds
# ------------------------------------------------------------------------------
build-linux-amd64: deps
	@mkdir -p $(BIN_DIR)/linux-amd64
	@echo "==> Building Linux x86_64 binary..."
	CGO_ENABLED=1 CC=$(CC_LINUX_AMD64) GOOS=linux GOARCH=amd64 \
	go build -trimpath -ldflags="$(LDFLAGS_LINUX)" \
		-o $(BIN_DIR)/linux-amd64/$(APP_NAME) ./cmd/vibeterm
	@echo "==> Linux x86_64 build complete: $(BIN_DIR)/linux-amd64/$(APP_NAME)"

build-linux-arm64: deps
	@mkdir -p $(BIN_DIR)/linux-arm64
	@echo "==> Cross-compiling Linux ARM64 binary..."
	CGO_ENABLED=1 CC=$(CC_LINUX_ARM64) GOOS=linux GOARCH=arm64 \
	go build -trimpath -ldflags="$(LDFLAGS_LINUX)" \
		-o $(BIN_DIR)/linux-arm64/$(APP_NAME) ./cmd/vibeterm
	@echo "==> Linux ARM64 build complete: $(BIN_DIR)/linux-arm64/$(APP_NAME)"

package-linux-amd64: build-linux-amd64
	@mkdir -p $(RELEASE_DIR)
	@echo "==> Packaging Linux x86_64 archive..."
	tar -czf $(RELEASE_DIR)/$(APP_NAME)-v$(VERSION)-linux-amd64.tar.gz \
		-C $(BIN_DIR)/linux-amd64 $(APP_NAME) \
		-C $(CURDIR) README.md
	@echo "==> Created $(RELEASE_DIR)/$(APP_NAME)-v$(VERSION)-linux-amd64.tar.gz"

package-linux-arm64: build-linux-arm64
	@mkdir -p $(RELEASE_DIR)
	@echo "==> Packaging Linux ARM64 archive..."
	tar -czf $(RELEASE_DIR)/$(APP_NAME)-v$(VERSION)-linux-arm64.tar.gz \
		-C $(BIN_DIR)/linux-arm64 $(APP_NAME) \
		-C $(CURDIR) README.md
	@echo "==> Created $(RELEASE_DIR)/$(APP_NAME)-v$(VERSION)-linux-arm64.tar.gz"

# ------------------------------------------------------------------------------
# 🪟 Windows Builds (MinGW-w64 + Embedded Resource Object)
# ------------------------------------------------------------------------------
windows-resource:
	@if command -v $(WINDRES) >/dev/null 2>&1; then \
		echo "==> Compiling Windows application manifest via windres..."; \
		$(WINDRES) -i cmd/vibeterm/vibeterm.rc -O coff -o cmd/vibeterm/vibeterm.syso 2>/dev/null || true; \
	fi

build-windows-amd64: deps windows-resource
	@mkdir -p $(BIN_DIR)/windows-amd64
	@echo "==> Cross-compiling Windows x86_64 binary via MinGW..."
	CGO_ENABLED=1 CC=$(CC_WINDOWS_AMD64) GOOS=windows GOARCH=amd64 \
	go build -trimpath -ldflags="$(LDFLAGS_WINDOWS)" \
		-o $(BIN_DIR)/windows-amd64/$(APP_NAME).exe ./cmd/vibeterm
	@rm -f cmd/vibeterm/vibeterm.syso
	@echo "==> Windows x86_64 build complete: $(BIN_DIR)/windows-amd64/$(APP_NAME).exe"

package-windows-amd64: build-windows-amd64
	@mkdir -p $(RELEASE_DIR)
	@echo "==> Packaging Windows x86_64 archive..."
	zip -j -9 $(RELEASE_DIR)/$(APP_NAME)-v$(VERSION)-windows-amd64.zip \
		$(BIN_DIR)/windows-amd64/$(APP_NAME).exe README.md
	@echo "==> Created $(RELEASE_DIR)/$(APP_NAME)-v$(VERSION)-windows-amd64.zip"

# ------------------------------------------------------------------------------
# 📦 Full Release & Checksum Generation
# ------------------------------------------------------------------------------
release: clean package-linux-amd64 package-windows-amd64 checksums
	@echo "================================================================================"
	@echo " ✅ All release packages built successfully in $(RELEASE_DIR):"
	@ls -lh $(RELEASE_DIR)
	@echo "================================================================================"

checksums:
	@echo "==> Computing SHA256 checksums..."
	@cd $(RELEASE_DIR) && sha256sum *.tar.gz *.zip > checksums.txt 2>/dev/null || true
	@if [ -f $(RELEASE_DIR)/checksums.txt ]; then \
		echo "==> Checksums generated:"; \
		cat $(RELEASE_DIR)/checksums.txt; \
	fi

# ------------------------------------------------------------------------------
# 🐳 Deterministic Containerized Multi-Stage Build
# ------------------------------------------------------------------------------
docker-build:
	@mkdir -p $(RELEASE_DIR)
	@echo "==> Building multi-platform release inside deterministic Docker container..."
	docker build -f cross-build.Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--target artifacts \
		--output type=local,dest=$(RELEASE_DIR) .
	@echo "==> Release artifacts extracted to $(RELEASE_DIR)"

run:
	go run ./cmd/vibeterm
