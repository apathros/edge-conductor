#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

MKFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TOP_DIR := $(MKFILE_DIR)/../..
PROTO_DIR := api/proto
PROTOBUF_VERSION := 3.14.0
PROTOC_FILENAME := protoc-$(PROTOBUF_VERSION)-linux-x86_64.zip

install-protoc:
	@if [ ! -e $(GOPATH)/bin/protoc ]; then \
		cd $$(mktemp -d) \
		&& wget -O protoc.zip https://github.com/google/protobuf/releases/download/v$(PROTOBUF_VERSION)/$(PROTOC_FILENAME) \
		&& unzip protoc.zip \
		&& mv bin/protoc $(GOPATH)/bin ;\
	fi
build:
	@if [ ! -e $(GOPATH)/bin/protoc ]; then \
		cd $$(mktemp -d) \
		&& wget -O protoc.zip https://github.com/google/protobuf/releases/download/v$(PROTOBUF_VERSION)/$(PROTOC_FILENAME) \
		&& unzip protoc.zip \
		&& mv bin/protoc $(GOPATH)/bin ;\
	fi
	@if [ ! -e $(GOPATH)/bin/protoc-gen-go ]; then \
		echo download protoc-gen-go \
		&& GO111MODULE=on go get -v -u github.com/golang/protobuf/proto@v1.4.3 \
		&& GO111MODULE=on go get -v -u github.com/golang/protobuf/protoc-gen-go@v1.4.3; \
	fi
	@cd $(TOP_DIR) && find $(PROTO_DIR) -name "*.proto" -exec protoc -I. --go_out=plugins=grpc:. {} \;
