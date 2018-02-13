VERSION_REF := $(shell git rev-parse --short HEAD)
VERSION_NAME := $(shell git describe --all | sed "s/^heads\///")

build:
	go build -x \
		-o build/graylog-proxy \
		-v -x \
		-ldflags '-w -s -X main.version=$(VERSION_NAME) -X main.ref=$(VERSION_REF)' \
		main.go && \
	strip build/graylog-proxy

clear:
	rm -rf build