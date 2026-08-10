# ── SubTracker Makefile ───────────────────────────────────────────────────────
BINARY  := subtracker
VERSION := 1.0.0
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
DIST    := dist

.PHONY: all build clean release tidy vet test help

## help: Show this help message
help:
	@echo ""
	@echo "  SubTracker v$(VERSION) — Build Commands"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""

## all: Tidy, vet, and build for current OS
all: tidy vet build

## tidy: Download and tidy Go module dependencies
tidy:
	go mod tidy

## vet: Run static analysis
vet:
	go vet ./...

## test: Run unit tests
test:
	go test -v ./...

## build: Build binary for the current OS/arch
build:
	go build $(LDFLAGS) -o $(BINARY) .

## release: Cross-compile for Linux, Windows, and macOS (amd64 + arm64)
release: tidy vet
	@mkdir -p $(DIST)

	@echo "  → linux/amd64"
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64       .

	@echo "  → linux/arm64"
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64        .

	@echo "  → darwin/amd64 (Intel Mac)"
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64       .

	@echo "  → darwin/arm64 (Apple Silicon)"
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64        .

	@echo "  → windows/amd64"
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe  .

	@echo ""
	@echo "  ✔  Binaries written to $(DIST)/"
	@ls -lh $(DIST)/

## clean: Remove built binaries and dist directory
clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -rf $(DIST)/
