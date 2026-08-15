# ==============================================================================
# VibeTerm Deterministic Multi-Platform Release Builder
# Multi-Stage Dockerfile with Linux & Windows CGO Toolchains
# ==============================================================================

# ------------------------------------------------------------------------------
# Stage 1: Build Environment (Pre-loaded with CGO Cross-Compilers & OpenGL Headers)
# ------------------------------------------------------------------------------
FROM golang:1.25-bookworm AS builder

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV GOTOOLCHAIN=auto

# Install native Linux graphics development headers, MinGW-w64 for Windows, and ARM64 cross-gcc
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    zip \
    tar \
    ca-certificates \
    git \
    # Linux Desktop / OpenGL / X11 / Wayland development headers
    libgl1-mesa-dev \
    xorg-dev \
    libwayland-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    libxkbcommon-dev \
    # Windows MinGW Cross-Compiler Suite
    gcc-mingw-w64-x86-64 \
    binutils-mingw-w64-x86-64 \
    mingw-w64-tools \
    # Linux ARM64 Cross-Compiler Suite
    gcc-aarch64-linux-gnu \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# Cache Go modules layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .
RUN go mod tidy

# Build Arguments
ARG VERSION=1.0.0
ARG GIT_COMMIT=release
ARG BUILD_TIME=2026-08-15T00:00:00Z

ENV LDFLAGS_BASE="-s -w -X 'vibeterm/internal/config.Version=${VERSION}' -X 'vibeterm/internal/config.GitCommit=${GIT_COMMIT}' -X 'vibeterm/internal/config.BuildTime=${BUILD_TIME}'"

# Output directories
RUN mkdir -p /out/bin /out/release

# 1. Compile Linux x86_64 Native Binary
RUN echo "==> Compiling Linux x86_64..." && \
    CGO_ENABLED=1 CC=gcc GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="${LDFLAGS_BASE}" \
    -o /out/bin/vibeterm-linux-amd64 ./cmd/vibeterm

# 2. Compile Windows x86_64 Native GUI Binary (.exe)
RUN echo "==> Compiling Windows x86_64 (.exe)..." && \
    x86_64-w64-mingw32-windres -i cmd/vibeterm/vibeterm.rc -O coff -o cmd/vibeterm/vibeterm.syso 2>/dev/null || true && \
    CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
    go build -trimpath -ldflags="${LDFLAGS_BASE} -H=windowsgui" \
    -o /out/bin/vibeterm.exe ./cmd/vibeterm && \
    rm -f cmd/vibeterm/vibeterm.syso

# 3. Package Archives
RUN echo "==> Packaging Release Archives..." && \
    # Linux tar.gz
    tar -czf /out/release/vibeterm-v${VERSION}-linux-amd64.tar.gz \
        -C /out/bin vibeterm-linux-amd64 \
        -C /src README.md && \
    # Windows zip
    cd /out/bin && zip -j -9 /out/release/vibeterm-v${VERSION}-windows-amd64.zip \
        vibeterm.exe /src/README.md && \
    # Compute SHA256 Checksums
    cd /out/release && sha256sum *.tar.gz *.zip > checksums.txt

# ------------------------------------------------------------------------------
# Stage 2: Scratch Artifact Exporter (Zero runtime image, exports files via BuildKit)
# ------------------------------------------------------------------------------
FROM scratch AS artifacts
COPY --from=builder /out/release/ /
