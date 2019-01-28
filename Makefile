# Copyright 2019 The KUARD Authors.
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

# This is a pretty complicated Makefile that builds the go binary (in a
# container) and then automates packaging it up into an image and pushing it. It
# then allows you to do this across multiple architectures and "fake versions".
#
# There is a bunch of funkiness around creating volumes so that intermetiate
# files (such as go libraries and npm module downloads) are cached across builds
# to speed things up.
#
# Some of the ideas here are taken from
# https://github.com/thockin/go-build-template and
# https://github.com/bowei/go-build-template.

# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

# Golang package.
PKG := github.com/kubernetes-up-and-running/kuard

# Registry to push to.
REGISTRY ?= gcr.io/kuar-demo

# For demo purposes, we want to build multiple versions.  They will all be
# mostly the same but will let us demonstrate rollouts.
FAKEVER ?= blue
ALL_FAKEVER = blue green purple

# This is the real version.  We'll grab it from git and use tags.
VERSION_BASE ?= $(shell git describe --tags --always --dirty)

# Set to 1 to print more verbose output from the build.
export VERBOSE ?= 0

# Default architecture to build for.
ARCH ?= amd64

ALL_ARCH := amd64 arm arm64 ppc64le
# Set default base image dynamically for each arch
ifeq ($(ARCH),amd64)
	BASEIMAGE?=alpine
endif
ifeq ($(ARCH),arm)
	BASEIMAGE?=arm32v6/alpine
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=arm64v8/alpine
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/alpine
endif

BUILD_IMAGE := kuard-build

DOCKER_RUN_FLAGS := --rm
DOCKER_BUILD_FLAGS := --rm
ifeq ($(VERBOSE), 1)
	VERBOSE_OUTPUT := >&1
else
	DOCKER_BUILD_FLAGS += -q
	VERBOSE_OUTPUT := >/dev/null
	MAKEFLAGS += -s
endif

DOCKER_MOUNTS:= \
	-v $(BUILD_IMAGE)-data:/data:delegated \
	-v $(BUILD_IMAGE)-node:/data/go/src/$(PKG)/client/node_modules:delegated \
	-v $$(pwd):/data/go/src/$(PKG):delegated \
	-v $$(pwd)/build:/build:delegated \
	-v $$(pwd)/bin/$(FAKEVER)/$(ARCH):/data/go/bin:delegated \
	-v $$(pwd)/bin/$(FAKEVER)/$(ARCH):/data/go/bin/linux_$(ARCH):delegated

DOCKER_ENVS:= \
	-e VERBOSE=$(VERBOSE)                                                \
	-e ARCH=$(ARCH)                                                      \
	-e PKG=$(PKG)                                                        \
	-e VERSION=$(VERSION_BASE)-$(FAKEVER)                                \

##############################################################################
# Default rule
all: build

##############################################################################
# Build container image

# Build the build image. This depends on the dockerfile for this image and the
# "init_data" script that initializes some volumes.  We use a "timestamp" to
# keep track of when this image was built and mirror the state of docker into
# the filesystem.
BUILD_IMAGE_BUILDSTAMP := .$(subst .,_,$(BUILD_IMAGE))-image
$(BUILD_IMAGE_BUILDSTAMP): build/init_data.sh Dockerfile.build
	@echo "container image: $(BUILD_IMAGE)"
	@echo "  Building container image"
	docker build                                                    \
		$(DOCKER_BUILD_FLAGS)                                         \
		-t $(BUILD_IMAGE)                                             \
		--build-arg "ALL_ARCH=$(ALL_ARCH)"                            \
		-f Dockerfile.build .                                         \
		$(VERBOSE_OUTPUT)
	@echo "  Creating volume $(BUILD_IMAGE)-data"
	-docker volume rm $(BUILD_IMAGE)-data $(VERBOSE_OUTPUT) 2>&1
	docker volume create $(BUILD_IMAGE)-data $(VERBOSE_OUTPUT)
	@echo "  Creating volume $(BUILD_IMAGE)-node"
	-docker volume rm $(BUILD_IMAGE)-node $(VERBOSE_OUTPUT) 2>&1
	docker volume create $(BUILD_IMAGE)-node $(VERBOSE_OUTPUT)
	@echo "  Running build/init_data.sh in build container to init volumes"
	docker run $(DOCKER_RUN_FLAGS)                   \
			-v $(BUILD_IMAGE)-data:/data:delegated       \
			-v $(BUILD_IMAGE)-node:/data/go/src/$(PKG)/client/node_modules:delegated \
			-v $$(pwd)/build:/build:delegated            \
			-e TARGET_UIDGID=$$(id -u):$$(id -g)         \
			$(BUILD_IMAGE)                               \
			/build/init_data.sh                          \
			$(VERBOSE_OUTPUT)
	echo "$(BUILD_IMAGE)" > $@
	docker images -q $(BUILD_IMAGE) >> $@

build-env: $(BUILD_IMAGE_BUILDSTAMP)
	@echo "Launching into build environment"
	docker run -ti                                                           \
			$(DOCKER_RUN_FLAGS)                                                  \
			$(DOCKER_MOUNTS)                                                     \
			--sig-proxy=true                                                     \
			$(DOCKER_ENVS)                                                       \
			-u $$(id -u):$$(id -g)                                               \
			-w /data/go/src/$(PKG)                                               \
			$(BUILD_IMAGE)                                                       \
			ash

##############################################################################
# Build the kuard binary
BINARYPATH:=bin/$(FAKEVER)/$(ARCH)/kuard

.PHONY: build
build: $(BINARYPATH)

$(BINARYPATH): build/build.sh $(BUILD_IMAGE_BUILDSTAMP)
	@echo "building binary: $@"
	@mkdir -p $(shell pwd)/bin/$(FAKEVER)/$(ARCH)
	docker run                                                               \
			$(DOCKER_RUN_FLAGS)                                                  \
			$(DOCKER_MOUNTS)                                                     \
			--sig-proxy=true                                                     \
			$(DOCKER_ENVS)                                                       \
			-u $$(id -u):$$(id -g)                                               \
			-w /data/go/src/$(PKG)                                               \
			$(BUILD_IMAGE)                                                       \
			./build/build.sh $(VERBOSE_OUTPUT)

##############################################################################
# Build the final container image

# Dockerfile for the final image. We use SED to slam a bunch of things in there.
BIN_DOCKERFILE:=.kuard-$(ARCH)-$(FAKEVER)-dockerfile
$(BIN_DOCKERFILE): Dockerfile.kuard
	@echo "generating Dockerfile $@ from $<"
	sed       \
			-e 's|ARG_ARCH|$(ARCH)|g'         \
			-e 's|ARG_FROM|$(BASEIMAGE)|g'    \
			-e 's|ARG_FAKEVER|$(FAKEVER)|g'   \
			$< > $@


CONTAINER_NAME  := $(REGISTRY)/kuard-$(ARCH)
BUILDSTAMP_NAME := $(subst /,_,$(CONTAINER_NAME)-$(FAKEVER))

.$(BUILDSTAMP_NAME)-image: $(BIN_DOCKERFILE) $(BINARYPATH)
	@echo "container image: $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER)"
	docker build                                                    \
		$(DOCKER_BUILD_FLAGS)                                         \
		-t $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER)               \
		-f .kuard-$(ARCH)-$(FAKEVER)-dockerfile .                     \
		$(VERBOSE_OUTPUT)
	echo "$(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER)" > $@
	@echo "container image tag: $(CONTAINER_NAME):$(FAKEVER)"
	docker tag $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER) $(CONTAINER_NAME):$(FAKEVER)
	echo "$(CONTAINER_NAME):$(FAKEVER)" >> $@
	docker images -q $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER) >> $@

.PHONY: images
images: .$(BUILDSTAMP_NAME)-image

##############################################################################
# Push to the registry

PUSH_BUILDSTAMP:=.$(BUILDSTAMP_NAME)-push

.PHONY: push
push: $(PUSH_BUILDSTAMP)

.%-push: .%-image
	@echo "pushing image: " $$(sed -n '1p' $<)
	docker push $$(sed -n '1p' $<) $(VERBOSE_OUTPUT)
	@echo "pushing image: " $$(sed -n '2p' $<)
	docker push $$(sed -n '2p' $<) $(VERBOSE_OUTPUT)
	cat $< > $@

##############################################################################
# Rules for dealing with fake versions
build-fakever-%:
	$(MAKE) --no-print-directory FAKEVER=$* build

images-fakever-%:
	$(MAKE) --no-print-directory FAKEVER=$* images

push-fakever-%:
	$(MAKE) --no-print-directory FAKEVER=$* push

.PHONY: all-fakever-build
all-fakever-build: $(addprefix build-fakever-, $(ALL_FAKEVER))

.PHONY: all-fakever-containers
all-fakever-containers: $(addprefix containers-fakever-, $(ALL_FAKEVER))

.PHONY: all-fakever-push
all-fakever-push: $(addprefix push-fakever-, $(ALL_FAKEVER))

##############################################################################
# Rules for dealing with multiple/all architectures at once

build-arch-%:
	$(MAKE) --no-print-directory ARCH=$* build

images-arch-%:
	$(MAKE) --no-print-directory ARCH=$* images

push-arch-%:
	$(MAKE) --no-print-directory ARCH=$* push

.PHONY: all-arch-build
all-arch-build: $(addprefix build-arch-, $(ALL_ARCH))

.PHONY: all-arch-containers
all-arch-containers: $(addprefix containers-arch-, $(ALL_ARCH))

.PHONY: all-arch-push
all-arch-push: $(addprefix push-arch-, $(ALL_ARCH))

##############################################################################
# Deal with all fakevers, all archs

.PHONY: all-build
all-build:
	@$(foreach ARCH,$(ALL_ARCH),\
		$(foreach FAKEVER,$(ALL_FAKEVER),\
			$(MAKE) --no-print-directory ARCH=$(ARCH) FAKEVER=$(FAKEVER) build;))

.PHONY: all-images
all-images:
	@$(foreach ARCH,$(ALL_ARCH),\
		$(foreach FAKEVER,$(ALL_FAKEVER),\
			$(MAKE) --no-print-directory ARCH=$(ARCH) FAKEVER=$(FAKEVER) images;))

.PHONY: all-push
all-push:
	@$(foreach ARCH,$(ALL_ARCH),\
		$(foreach FAKEVER,$(ALL_FAKEVER),\
			$(MAKE) --no-print-directory ARCH=$(ARCH) FAKEVER=$(FAKEVER) push;))

##############################################################################
# Misc commands
.PHONY: version
version:
	@echo $(VERSION_BASE)

.PHONY: clean
clean: container-clean bin-clean

.PHONY: container-clean
container-clean:
	docker volume rm -f $(BUILD_IMAGE)-data $(BUILD_IMAGE)-node $(VERBOSE_OUTPUT)
	rm -f .*-container .*-dockerfile .*-push .*-image

.PHONY: bin-clean
bin-clean:
	rm -rf bin

.PHONY: client-clean
	rm -rf client/node_modules sitedata/built

.PHONY: help
help:
	@echo "make targets"
	@echo
	@echo "  all, build    build all binaries"
	@echo "  images        build the container image"
	@echo "  push          push images to the registry"
	@echo "  clean         clean up all files and docker volumes/images"
	@echo "  help          this help message"
	@echo "  version       show package version"
	@echo
	@echo "  {build,images,push}-arch-ARCH    do action for specific ARCH"
	@echo "  all-arch-{build,images,push}     do action for all arches"
	@echo
	@echo "  {build,images,push}-fakever-FAKEVER  do action for specific FAKEVER"
	@echo "  all-fakever-{build,images,push}      do action for all fakevers"
	@echo
	@echo "  all-{build,images,push}    do action fo all arches and all fakevers"
	@echo
	@echo "  Available ARCH: $(ALL_ARCH)"
	@echo "  Default FAKEVERS: $(ALL_FAKEVER)"
	@echo
	@echo "  Setting VERBOSE=1 will show additional build logging."
	@echo
	@echo "  Setting VERSION_BASE will override the container version tag."
