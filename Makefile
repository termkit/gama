.PHONY: all linux windows macos_amd64 macos_arm64 build get_latest_tag

LATEST_TAG ?= $(shell git describe --tags `git rev-list --tags --max-count=1`)

all: get_latest_tag linux windows macos_amd64 macos_arm64

linux: get_latest_tag
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-linux-amd64 main.go

windows: get_latest_tag
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-windows-amd64.exe main.go

macos_amd64: get_latest_tag
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-macos-amd64 main.go

macos_arm64: get_latest_tag
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-macos-arm64 main.go

build: get_latest_tag # build for current OS
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-$(GOOS)-$(GOARCH) main.go

get_latest_tag:
	@echo "Getting latest Git tag..."
	@echo $(LATEST_TAG)