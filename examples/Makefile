#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#
.PHONY:clean all

SUBDIRS=$(shell ls -l | grep ^d)

build:
	@for dir in $(SUBDIRS) ; do \
        	if test -d $$dir ; then \
	       		echo "$$dir: $(MAKE)" ; \
			if (cd $$dir; $(MAKE)) ; then \
				true; \
			fi; \
		fi \
	done

