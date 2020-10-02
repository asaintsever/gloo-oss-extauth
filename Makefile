SHELL=/bin/bash

BIN:=gloo-oss-extauth

.SILENT: ;  	# No need for @
.ONESHELL: ; 	# Single shell for a target (required to properly use local variables)
.PHONY: clean build image
.DEFAULT_GOAL := build

clean:
	rm -f target/*

build: clean  # run 'make build OFFLINE=true' to build from vendor folder
	if [ -z ${OFFLINE} ] || [ ${OFFLINE} != true ];then \
		echo "Building ..."; \
		GOOS=linux GOARCH=amd64 go build -mod=mod -a -o target/$(BIN) ./cmd; \
	else \
		echo "Building using local vendor folder (ie offline build) ..."; \
		GOOS=linux GOARCH=amd64 go build -mod=vendor -a -o target/$(BIN) ./cmd; \
	fi

image:
	echo "Build image using Go container and multi-stage build ..."
	docker build -t asaintsever/$(BIN) .
