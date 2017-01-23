#!/bin/bash
#
# Copyright 2016 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail

echo "Verbose: ${VERBOSE}"
GO_FLAGS=
if [[ "${VERBOSE:-0}" = "1" ]]; then
  echo "Building with VERBOSE"
  GO_FLAGS="-x"
  set -o xtrace
fi

if [ -z "${PKG}" ]; then
    echo "PKG must be set"
    exit 1
fi
if [ -z "${ARCH}" ]; then
    echo "ARCH must be set"
    exit 1
fi
if [ -z "${VERSION}" ]; then
    echo "VERSION must be set"
    exit 1
fi

export CGO_ENABLED=0
export GOARCH="${ARCH}"

(
  cd client
  npm install --loglevel=error
  npm run build
)

go generate ${GO_FLAGS} ./cmd/... ./pkg/...
go install                                                         \
    ${GO_FLAGS}                                                    \
    -installsuffix "static"                                        \
    -ldflags "-X ${PKG}/pkg/version.VERSION=${VERSION}"            \
    ./cmd/...
