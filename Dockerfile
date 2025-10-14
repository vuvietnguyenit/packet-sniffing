# =========================================
# 1️⃣ Build stage — compile eBPF + Go binary
# =========================================
FROM golang:1.25-1-bookworm AS builder

# Install clang, llvm, bpftool (for bpf2go to work)
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang llvm libbpf-dev bpftool \
    && rm -rf /var/lib/apt/lists/*

# Set workdir
WORKDIR /app

# Copy go.mod and download deps first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Generate eBPF bindings (runs bpf2go)
RUN go generate ./...

# Build Go binary
RUN CGO_ENABLED=0 go build -o /app/traceapp .

# =========================================
# 2️⃣ Runtime stage — minimal container
# =========================================
FROM debian:bookworm-slim

# Needed for kernel interaction and debugging
RUN apt-get update && apt-get install -y --no-install-recommends bpftool iproute2 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/traceapp .

# eBPF programs require privileged mode & access to /sys
CMD ["./traceapp"]
