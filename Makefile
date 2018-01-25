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

#
# `make help` will show commonly used targets.
#

# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

# Golang package.
PKG := github.com/kubernetes-up-and-running/kuard

# List of binaries to build. You must have a matching Dockerfile.BINARY
# for each BINARY.
BINARIES := kuard

# Registry to push to.
REGISTRY ?= gcr.io/kuar-demo

# Default architecture to build for.
ARCH ?= amd64

# Image to use for building.
BUILD_IMAGE ?= kuard-build

# For demo purposes, we want to build multiple versions.  They will all be
# mostly the same but will let us demonstrate rollouts.
FAKE_VERSIONS = 1 2 3

# This is the real version.  We'll grab it from git and use tags.
VERSION_BASE ?= $(shell git describe --tags --always --dirty)
# This version-strategy uses a manual value to set the version string
#VERSION := 1.2.3

DOCKER_MOUNT_OPTION ?= :delegated

# Set to 1 to print more verbose output from the build.
export VERBOSE ?= 0

# Include standard build rules.
include rules.mk
