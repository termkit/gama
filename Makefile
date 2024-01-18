.PHONY: all linux_amd64 windows macos_amd64 macos_arm64 build get_latest_tag

LATEST_TAG ?= $(shell git describe --tags `git rev-list --tags --max-count=1`)

all: get_latest_tag linux_amd64 linux_arm64 windows macos_amd64 macos_arm64

linux_amd64: get_latest_tag
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-linux-amd64 main.go

linux_arm64: get_latest_tag
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-linux-arm64 main.go

windows: get_latest_tag
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-windows-amd64.exe main.go

macos_amd64: get_latest_tag
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-macos-amd64 main.go

macos_arm64: get_latest_tag
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama-macos-arm64 main.go

build: get_latest_tag # build for current OS
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(LATEST_TAG)" -o release/gama main.go

get_latest_tag:
	@echo "Getting latest Git tag..."
	@echo $(LATEST_TAG)