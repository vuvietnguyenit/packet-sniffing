ARCH := x86
BINARY := mysql-error-echo
BUILD_DIR := ./bin
APP_DIR := ./app
GOSRC     := ./$(APP_DIR)/cmd ./$(APP_DIR)/internal ./$(APP_DIR)/*.go

# Go related variables
GO      ?= go
GOFLAGS :=
LDFLAGS := -s -w

# Get Linux kernel version by integer value


.PHONY: all build run clean bpf-gen

# Default target
all: build

# Generate eBPF Go code via go:generate
BPF_CLANG ?= clang
		
BPF_CFLAGS := -O2 -g -Wall -target bpf -D__TARGET_ARCH_x86  -Wno-error=unknown-warning-option -Wno-address-of-packed-member -Wno-unused-value -ferror-limit=0 
BPF_SRC := app/bpf/mysql_response_trace.c
BPF_OBJ := app/bpf/mysql_response_trace.bpf.o
BPF_VMLINUX := app/bpf/vmlinux.h

bpf-gen:
	@echo ">> Generating eBPF Go code using go:generate"
	$(BPF_CLANG) $(BPF_CFLAGS) \
		-I$(dir $(BPF_VMLINUX)) \
		-c $(BPF_SRC) -o $(BPF_OBJ)


bpf-gen-btf:
	@echo ">> generate vmlinux.h using bpftool"
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > ./$(APP_DIR)/bpf/vmlinux.h

# Build the Go binary
build: bpf-gen-btf bpf-gen
	@echo ">> Building $(BINARY)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./app

# Run the binary
run: build
	@echo ">> Running $(BINARY)"
	$(BUILD_DIR)/$(BINARY)

# Clean generated files
clean:
	rm -f bpf_$(ARCH)_*.o bpf_$(PROG)_*.go
	rm -rf $(BINARY)

# Build docker
docker-build:
	docker build -t $(ARGS) .

docker-push:
	docker push $(ARGS)