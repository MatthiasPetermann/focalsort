# ============================================================================
# Focalsort Makefile 
# ============================================================================

# Name of the compiled binary
BINARY = focalsort

# Define the version for this build
VERSION := v0.1.0

# Path where the final binary will be written
OUT = ./bin/$(BINARY)

# Default target: builds Lunaria using the selected profile
all: build

build: 
	go build -ldflags="-X 'd2ux.net/focalsort/internal/core.Version=$(VERSION)'" \
		-o $(OUT) 

# ============================================================================
# build-stripped – Same as `build`, but with stripped symbols for a smaller binary
# This removes debugging and DWARF info; great for production binaries
# ============================================================================
build-stripped: 
	go build -ldflags="-s -w -X 'd2ux.net/focalsort/internal/core.Version=$(VERSION)'" \
		-o $(OUT)

# ============================================================================
# build-static – Builds a fully static binary (Linux/AMD64) with CGO disabled
# Injects version and strips symbols.
# ============================================================================
build-static: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags="-s -w -extldflags '-static' -X 'd2ux.net/focalsort/internal/core.Version=$(VERSION)'" \
		-o $(OUT) 

# ============================================================================
# clean – Removes the build output directory (./bin)
# ============================================================================
clean:
	rm -rf ./bin

# ============================================================================
# fmt – Applies Go formatting to all source files
# ============================================================================
fmt:
	go fmt ./...

# ============================================================================
# tidy – Cleans up go.mod and go.sum (removes unused dependencies)
# ============================================================================
tidy:
	go mod tidy

# ============================================================================
# test – Runs all Go tests in the project
# ============================================================================
test:
	go test ./...

# ============================================================================
# Declare all non-file targets as phony
# ============================================================================
.PHONY: all build build-stripped build-static clean fmt tidy test print-tags list-profiles generate_imports check-tools

