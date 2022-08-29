#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

WORKSPACE = _workspace
BINDIR = $(WORKSPACE)/bin
BINNAME = conductor
PKI = $(WORKSPACE)/cert/pki
GOVERSION = 1.17-stretch
KUBECTL_VERSION = v1.20.0
GOLANGCI_LINT_VERSION = 1.44.0
GODIR = .go
GO_BUILD_MODE=-ldflags '-linkmode=external' -buildmode=pie
GO_BUILD_CMD=docker run --rm --hostname=ctn-$(shell hostname) --network host --user $(shell id -u):$(shell id -g)\
	-e http_proxy="$(http_proxy)" -e https_proxy="$(https_proxy)" -e no_proxy="$(no_proxy)" \
	-e HTTP_PROXY="$(HTTP_PROXY)" -e HTTPS_PROXY="$(HTTPS_PROXY)" -e NO_PROXY="$(NO_PROXY)" \
	-e GOCACHE=/go/.cache \
	-v $(PWD):$(PWD) \
	-v $(PWD)/$(GODIR):/go \
	-t \
	golang:$(GOVERSION)

.PHONY: build docker-build prebuild gen-code gen-hash test docker-golangci-lint gen-doc
all: docker-build docker-golangci-lint gen-doc

prebuild:
	mkdir -p $(WORKSPACE)
	mkdir -p $(PWD)/$(GODIR)/bin
	make -C api/proto install-protoc GOPATH=$(PWD)/$(GODIR)

gen-code: prebuild
	mkdir -p $(GODIR)/.cache
	mkdir -p $(BINDIR)
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && make host-code-generate"
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && make host-code-format"

gen-doc: gen-code
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && go run build/errormdgenerator.go"

build: gen-code
	mkdir -p $(GODIR)/.cache
	mkdir -p $(BINDIR)
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && make host-build"

build_shell:
	mkdir -p $(GODIR)/.cache
	mkdir -p $(BINDIR)
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && bash"

gen-hash: prebuild
	mkdir -p $(WORKSPACE)/config
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && go mod tidy && go run build/hashgenerator.go"

host-code-generate: prebuild
	@rm -rf api/.schemas
	@cp -r api/schemas api/.schemas
	make -C api/.schemas
	make -C api/proto build
	go mod tidy
	go run build/plugingenerator.go
	@rm -rf api/.schemas
	go vet ./pkg/... ./cmd/...

host-code-format: prebuild
	@echo "Going to format and build code in ${PWD}"
	@go fmt $(shell sh -c "go list ./...")
	@echo "--- Finished code format ---"

host-build: prebuild
	go build -v $(GO_BUILD_MODE) -o $(BINDIR)/$(BINNAME) cmd/ep/main.go
	rm -f $(WORKSPACE)/$(BINNAME) && ln -s bin/$(BINNAME) $(WORKSPACE)/$(BINNAME)
	go build -v $(GO_BUILD_MODE) -o $(BINDIR)/$(BINNAME)-plugin cmd/plugin/main.go
	test -f $(BINDIR)/kubectl || curl -fsSL https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(shell uname | tr '[:upper:]' '[:lower:]')/$(shell uname -m | tr '[:upper:]' '[:lower:]' | sed -e s/x86_64/amd64/)/kubectl -o $(BINDIR)/kubectl && chmod +x $(BINDIR)/kubectl
	rm -f $(WORKSPACE)/kubectl && ln -s bin/kubectl $(WORKSPACE)/kubectl
	rm -rf $(WORKSPACE)/config && cp -a configs $(WORKSPACE)/config && cp -a examples/* $(WORKSPACE)/config/
	rm -rf $(WORKSPACE)/kit && cp -a kit $(WORKSPACE)/kit && chmod -R 600 $(WORKSPACE)/kit/*
	rm -rf $(WORKSPACE)/services && cp -a services $(WORKSPACE)/services
	rm -rf $(WORKSPACE)/workflow && mv $(WORKSPACE)/config/workflow $(WORKSPACE)/workflow
	make -C $(WORKSPACE)/services build
	make -C $(WORKSPACE)/config build
	go run build/hashgenerator.go

docker-build: build

bash_completion: build
	cd $(WORKSPACE) && ./$(BINNAME) completion bash | sed "s#complete -o default -F __start_ep $(BINNAME)#complete -o default -F __start_ep \./$(BINNAME)#">~/.bash_completion
	cd -
clean:
	@( if [ -e $(GODIR) ]; then chmod +w $(GODIR) -R && rm -fr $(GODIR) ; fi )
	@( rm -f $(BINDIR)/$(BINNAME) )
	@( rm -f $(BINDIR)/$(BINNAME)-plugin )
	@( rm -rf $(PKI) )

test-build-dependency:
	go install github.com/golang/mock/mockgen@v1.6.0
	go generate ./...

test-unittest: test-build-dependency
	go test -gcflags=-l `go list ep/pkg/... ep/cmd/... | grep -v "/mock" | grep -v "/test/" | grep -v "/pkg/api"` -v -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	go tool cover -func cover.out -o cover.function-coverage.log

host-test: 
	mkdir -p $(PWD)/$(GODIR)
	make test-unittest

docker-test:
	mkdir -p $(PWD)/$(GODIR)
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && \
		make host-test"

docker-golangci-lint:
	mkdir -p $(PWD)/$(GODIR)/.cache
	mkdir -p $(PWD)/$(GODIR)
	$(GO_BUILD_CMD) bash -c "cd $(PWD) && \
		if [ ! -e .golangci.yml ] ; then exit 0 ; fi && \
		export GOLANGCI_LINT_CACHE=/go/.cache && \
		if [ ! -e $(PWD)/$(GODIR)/bin/golangci-lint ] || ! $(PWD)/$(GODIR)/bin/golangci-lint version | grep $(GOLANGCI_LINT_VERSION) ; then \
			curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PWD)/$(GODIR)/bin v$(GOLANGCI_LINT_VERSION); \
		fi && \
		$(PWD)/$(GODIR)/bin/golangci-lint run pkg/... cmd/... --config .golangci.yml"

test: docker-test
	
artifact: clean
	git clean -f -d -X
	tar --exclude=EdgeConductor-*.tar.gz --exclude=Jenkinsfile --exclude=copyright.cfg -cvzf EdgeConductor-open-`cat ./VERSION`.tar.gz *
