# Copyright 2017 The KUAR Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.7-alpine

# We aren't using the /go directory defined in base image
WORKDIR /data
ENV GOPATH /data/go

ENV npm_config_cache=/data/npm_cache

# Create links based on passed in ALL_ARCH
ARG ALL_ARCH
ENV ALL_ARCH ${ALL_ARCH}
RUN for ARCH in ${ALL_ARCH}; do \
      ln -s -f "/data/std/${ARCH}" "/usr/local/go/pkg/linux_${ARCH}_static" ; \
    done

RUN apk update && apk upgrade && apk add --no-cache git nodejs bash

# Install any binaries in to /usr/local/bin instead of /go/bin as we will mask
# /go/bin when doing real builds.
RUN GOPATH=/tmp GOBIN=/usr/local/bin go get -u github.com/jteeuwen/go-bindata/...
RUN GOPATH=/tmp GOBIN=/usr/local/bin go get github.com/tools/godep
