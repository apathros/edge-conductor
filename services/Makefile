#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

build:
	for dir in $$(ls | grep -iE "*-helm"); do \
		tar -zcvf $$dir.tgz $$dir; \
	done
