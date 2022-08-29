#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

SCHEMAS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
APIPKG_DIR := ../../../pkg/api
SWAGGER_VERSION := v0.28.0
SWAGGER_URL := https://github.com/go-swagger/go-swagger/releases/download/$(SWAGGER_VERSION)/swagger_linux_amd64

models := $(patsubst $(SCHEMAS_DIR)%/,%,$(filter-out $(SCHEMAS_DIR),$(dir $(wildcard $(SCHEMAS_DIR)*/))))

build:
	@if [ ! -e $(GOPATH)/bin/swagger ] || ! $(GOPATH)/bin/swagger version | grep $(SWAGGER_VERSION) ; then \
		echo download swagger \
		&& mkdir -p $(GOPATH)/bin \
		&& wget -O $(GOPATH)/bin/.swagger.dl "$(SWAGGER_URL)" \
		&& chmod +x $(GOPATH)/bin/.swagger.dl \
		&& mv $(GOPATH)/bin/.swagger.dl $(GOPATH)/bin/swagger; \
	fi
	@cd $(SCHEMAS_DIR) \
	&& chmod a+x pattern.sh \
	&& ./pattern.sh \
	&& for m in $(models); do \
		cd $$m > /dev/null \
		&& mkdir -p $(APIPKG_DIR)/$$m \
		&& for s in $$(ls *.yml); do \
			if ! grep "\$$ref:.*$$s" . -qRw; then \
				echo gen api from $$s; \
				$(GOPATH)/bin/swagger generate model -f $$s --accept-definitions-only -m $(APIPKG_DIR)/$$m -q -r ../license-header.txt || exit 1 ; \
			fi ; \
		done \
		&& cd - > /dev/null || exit 1; \
	done
