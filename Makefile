# Makefile for command-runner
BINARY=command-runner

# These are the values we want to pass for VERSION and BUILD
VERSION=`git describe --tags`
BUILD_DATE=`date +%FT%T%z`
USER=`git config user.email`
BRANCH=`git rev-parse --abbrev-ref HEAD`
REVISION=`git rev-parse --short HEAD`

# Setup the -ldflags option for go build here, interpolate the variable values.
# Note the Version module is in a different git repo.
MODULE="github.com/perforce/p4prometheus"
LOCAL_LDFLAGS=-ldflags="-X  ${MODULE}/version.Version=${VERSION} -X ${MODULE}/version.BuildDate=${BUILD_DATE} -X ${MODULE}/version.Branch=${BRANCH} -X ${MODULE}/version.Revision=${REVISION} -X ${MODULE}/version.BuildUser=${USER}"
LDFLAGS=-ldflags="-w -s -X ${MODULE}/version.Version=${VERSION} -X ${MODULE}/version.BuildDate=${BUILD_DATE} -X ${MODULE}/version.Branch=${BRANCH} -X ${MODULE}/version.Revision=${REVISION} -X ${MODULE}/version.BuildUser=${USER}"

.PHONY: build install clean

build:
	@echo "Building..."
	@go build $(LOCAL_LDFLAGS)

# Builds distributions for Windows, Mac, Linux
dist:
	GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY}-darwin-amd64 ${LDFLAGS}
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY}-linux-amd64 ${LDFLAGS}
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY}-windows-amd64.exe ${LDFLAGS}
	rm -f bin/${BINARY}*amd64*.gz
	rm bin/cmd_config.yaml
	cp ./configs/cmd_config.yaml bin/
	cd bin && tar czf ${BINARY}-linux-amd64.tar.gz ${BINARY}-linux-amd64 cmd_config.yaml
	-chmod +x bin/${BINARY}-darwin-amd64
	-chmod +x bin/${BINARY}-windows-amd64
	gzip bin/${BINARY}-darwin-amd64
	gzip bin/${BINARY}-windows-amd64




install:
	@echo "Installing..."
	@go install -ldflags "$(LDFLAGS)"

clean:
	@echo "Cleaning..."
	@rm -f command-runner
