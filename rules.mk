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
# These build rules should not need to be modified.
#
SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

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

# These rules MUST be expanded at reference time (hence '=') as BINARY
# is dynamically scoped.
CONTAINER_NAME  = $(REGISTRY)/$(BINARY)-$(ARCH)
BUILDSTAMP_NAME = $(subst /,_,$(CONTAINER_NAME)-$(FAKEVER))

# We need to build a separate set of binaries for each fake version.  We'll do
# this by putting them in a bin dir based on the FAKEVER.
GO_FAKEVER_BINARIES := $(foreach FAKEVER,$(FAKE_VERSIONS),\
  $(addprefix bin/$(FAKEVER)/$(ARCH)/,$(BINARIES)))
CONTAINER_BUILDSTAMPS := \
  $(foreach BINARY,$(BINARIES),\
	  $(foreach FAKEVER,$(FAKE_VERSIONS),\
		  .$(BUILDSTAMP_NAME)-container))
PUSH_BUILDSTAMPS := \
  $(foreach BINARY,$(BINARIES),\
	  $(foreach FAKEVER,$(FAKE_VERSIONS),\
		  .$(BUILDSTAMP_NAME)-push))
BUILD_IMAGE_BUILDSTAMP := .$(subst .,_,$(BUILD_IMAGE))-container

DOCKER_RUN_FLAGS := --rm
DOCKER_BUILD_FLAGS := --rm
ifeq ($(VERBOSE), 1)
	VERBOSE_OUTPUT := >&1
else
	DOCKER_BUILD_FLAGS += -q
	VERBOSE_OUTPUT := >/dev/null
	MAKEFLAGS += -s
endif

all: build

build-%:
	$(MAKE) --no-print-directory ARCH=$* build

containers-%:
	$(MAKE) --no-print-directory ARCH=$* containers

push-%:
	$(MAKE) --no-print-directory ARCH=$* push


.PHONY: all-build
all-build: $(addprefix build-, $(ALL_ARCH))

.PHONY: all-containers
all-containers: $(addprefix containers-, $(ALL_ARCH))

.PHONY: all-push
all-push: $(addprefix push-, $(ALL_ARCH))

.PHONY: build
build: $(GO_FAKEVER_BINARIES)

$(BUILD_IMAGE_BUILDSTAMP): build/init_data.sh Dockerfile.build
	@echo "container: $(BUILD_IMAGE)"
	docker build                                                    \
		$(DOCKER_BUILD_FLAGS)                                         \
		-t $(BUILD_IMAGE)                                             \
		--build-arg "ALL_ARCH=$(ALL_ARCH)"                            \
		-f Dockerfile.build .                                         \
		$(VERBOSE_OUTPUT)
	docker volume create $(BUILD_IMAGE)-data $(VERBOSE_OUTPUT)
	docker volume create $(BUILD_IMAGE)-node $(VERBOSE_OUTPUT)
	docker run $(DOCKER_RUN_FLAGS)                   \
	    -v $(BUILD_IMAGE)-data:/data$(DOCKER_MOUNT_OPTION) \
	    -v $(BUILD_IMAGE)-node:/data/go/src/$(PKG)/client/node_modules$(DOCKER_MOUNT_OPTION) \
	    -v $$(pwd)/build:/build$(DOCKER_MOUNT_OPTION) \
	    -e TARGET_UIDGID=$$(id -u):$$(id -g)         \
	    $(BUILD_IMAGE)                               \
	    /build/init_data.sh                          \
	    $(VERBOSE_OUTPUT)
	echo "$(BUILD_IMAGE)" > $@
	docker images -q $(BUILD_IMAGE) >> $@

# Rules for all bin/$(FAKEVER)/$(ARCH)/$(BINARY)
GO_BINARIES = $(addprefix bin/$(FAKEVER)/$(ARCH)/,$(BINARIES))
define GO_BINARIES_RULE
# Make this target phony so we always rebuild it. We don't track build
# dependencies.
.PHONY: $(GO_BINARIES)
$(GO_BINARIES): build/build.sh $(BUILD_IMAGE_BUILDSTAMP)
	@echo "building : $$@"
	docker run                                                               \
	    $(DOCKER_RUN_FLAGS)                                                  \
	    --sig-proxy=true                                                     \
	    -e VERBOSE=$(VERBOSE)                                                \
	    -e ARCH=$(ARCH)                                                      \
	    -e PKG=$(PKG)                                                        \
	    -e VERSION=$(VERSION_BASE)-$(FAKEVER)                                \
	    -u $$$$(id -u):$$$$(id -g)                                           \
	    -v $(BUILD_IMAGE)-data:/data$(DOCKER_MOUNT_OPTION)                   \
	    -v $(BUILD_IMAGE)-node:/data/go/src/$(PKG)/client/node_modules$(DOCKER_MOUNT_OPTION) \
	    -v $$$$(pwd):/data/go/src/$(PKG)$(DOCKER_MOUNT_OPTION)               \
	    -v $$$$(pwd)/bin/$(FAKEVER)/$(ARCH):/data/go/bin$(DOCKER_MOUNT_OPTION) \
	    -v $$$$(pwd)/bin/$(FAKEVER)/$(ARCH):/data/go/bin/linux_$(ARCH)$(DOCKER_MOUNT_OPTION) \
	    -w /data/go/src/$(PKG)                                               \
	    $(BUILD_IMAGE)                                                       \
	    ./build/build.sh $(VERBOSE_OUTPUT)
endef
$(foreach FAKEVER,$(FAKE_VERSIONS),\
  $(eval $(GO_BINARIES_RULE)))

# Rules for dockerfiles.
define DOCKERFILE_RULE
.$(BINARY)-$(ARCH)-$(FAKEVER)-dockerfile: Dockerfile.$(BINARY)
	@echo generating Dockerfile $$@ from $$<
	sed					\
	    -e 's|ARG_BIN|$(BINARY)|g'        \
	    -e 's|ARG_ARCH|$(ARCH)|g'         \
	    -e 's|ARG_FROM|$(BASEIMAGE)|g'    \
			-e 's|ARG_FAKEVER|$(FAKEVER)|g'   \
	    $$< > $$@
.$(BUILDSTAMP_NAME)-container: .$(BINARY)-$(ARCH)-$(FAKEVER)-dockerfile
endef
$(foreach BINARY,$(BINARIES),\
  $(foreach FAKEVER,$(FAKE_VERSIONS),\
	  $(eval $(DOCKERFILE_RULE))))


# Rules for containers
define CONTAINER_RULE
.$(BUILDSTAMP_NAME)-container: bin/$(FAKEVER)/$(ARCH)/$(BINARY)
	@echo "container: $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER) (bin/$(ARCH)/$(BINARY))"
	docker build                                                    \
		$(DOCKER_BUILD_FLAGS)                                         \
		-t $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER)               \
		-f .$(BINARY)-$(ARCH)-$(FAKEVER)-dockerfile .                 \
		$(VERBOSE_OUTPUT)
	echo "$(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER)" > $$@
	@echo "container: $(CONTAINER_NAME):$(FAKEVER) (bin/$(ARCH)/$(BINARY))"
	docker tag $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER) $(CONTAINER_NAME):$(FAKEVER)
	echo "$(CONTAINER_NAME):$(FAKEVER)" >> $$@
	docker images -q $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER) >> $$@
endef
$(foreach BINARY,$(BINARIES),\
  $(foreach FAKEVER,$(FAKE_VERSIONS),\
	  $(eval $(CONTAINER_RULE))))

.PHONY: containers
containers: $(CONTAINER_BUILDSTAMPS)


# Rules for pushing
.PHONY: push
push: $(PUSH_BUILDSTAMPS)

.%-push: .%-container
	@echo "pushing  :" $$(sed -n '1p' $<)
	gcloud docker -- push $$(sed -n '1p' $<) $(VERBOSE_OUTPUT)
	@echo "pushing  :" $$(sed -n '2p' $<)
	gcloud docker -- push $$(sed -n '2p' $<) $(VERBOSE_OUTPUT)
	cat $< > $@

define PUSH_RULE
only-push-$(BINARY): .$(BUILDSTAMP_NAME)-push
endef
$(foreach BINARY,$(BINARIES),\
  $(foreach FAKEVER,$(FAKE_VERSIONS),\
    $(eval $(PUSH_RULE))))

.PHONY: push-names
push-names:
	@$(foreach BINARY,$(BINARIES),\
	  $(foreach FAKEVER,$(FAKE_VERSIONS),\
	    echo $(CONTAINER_NAME):$(VERSION_BASE)-$(FAKEVER);\
			echo $(CONTAINER_NAME):$(FAKEVER);))


# Rule for `test`
.PHONY: test
test: $(BUILD_IMAGE_BUILDSTAMP)
	docker run                                                             \
	    $(DOCKER_RUN_FLAGS)                                                \
	    --sig-proxy=true                                                   \
	    -u $$(id -u):$$(id -g)                                             \
			-v $(BUILD_IMAGE)-data:/data$(DOCKER_MOUNT_OPTION)                 \
	    -v $(BUILD_IMAGE)-node:/data/go/src/$(PKG)/client/node_modules$(DOCKER_MOUNT_OPTION) \
	    -v $$(pwd):/data/go/src/$(PKG)$(DOCKER_MOUNT_OPTION)               \
	    -v $$(pwd)/bin/$(ARCH):/data/go/bin$(DOCKER_MOUNT_OPTION)          \
	    -w /data/go/src/$(PKG)                                             \
	    $(BUILD_IMAGE)                                                     \
	    /bin/sh -c "                                                       \
	        ./build/test.sh $(SRC_DIRS)                                    \
	    "

# Miscellaneous rules
.PHONY: version
version:
	@echo $(VERSION_BASE)

.PHONY: clean
clean: container-clean bin-clean

.PHONY: container-clean
container-clean:
	docker volume rm -f $(BUILD_IMAGE)-data $(BUILD_IMAGE)-node $(VERBOSE_OUTPUT)
	rm -f .*-container .*-dockerfile .*-push

.PHONY: bin-clean
bin-clean:
	rm -rf bin

.PHONY: help
help:
	@echo "make targets"
	@echo
	@echo "  all, build    build all binaries"
	@echo "  containers    build the containers"
	@echo "  push          push containers to the registry"
	@echo "  help          this help message"
	@echo "  version       show package version"
	@echo
	@echo "  {build,containers,push}-ARCH    do action for specific ARCH"
	@echo "  all-{build,containers,push}     do action for all ARCH"
	@echo "  only-push-BINARY                push just BINARY"
	@echo "  push-names                      print all of the container names"
	@echo
	@echo "  Available ARCH: $(ALL_ARCH)"
	@echo "  Available BINARIES: $(BINARIES)"
	@echo
	@echo "  Setting VERBOSE=1 will show additional build logging."
	@echo
	@echo "  Setting VERSION_BASE will override the container version tag."
