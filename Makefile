VERSION := 1.0.11-installtest
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date +%F-%T)
LDFLAGS := -X main.version=$(VERSION)-$(COMMIT_HASH)-$(BUILD_DATE)

.PHONY: build install clean

build:
	@echo "Building..."
	@go build -ldflags "$(LDFLAGS)" -o command-runner

install:
	@echo "Installing..."
	@go install -ldflags "$(LDFLAGS)"

clean:
	@echo "Cleaning..."
	@rm -f command-runner
