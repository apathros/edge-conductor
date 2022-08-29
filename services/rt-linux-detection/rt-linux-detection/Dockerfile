#
# Copyright (c) 2022 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#
#
FROM bitnami/kubectl:1.20.9 as kubectl
FROM alpine:latest

# copy kubectl  
COPY --from=kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/
RUN mkdir -p /scripts
COPY rt_linux_detect.sh /scripts
