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
	BASEIMAGE?=armel/busybox
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=aarch64/busybox
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/busybox
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

$(BUILD_IMAGE_BUILDSTAMP): Dockerfile.build
	@echo "container: $(BUILD_IMAGE)"
	docker build                                                    \
		$(DOCKER_BUILD_FLAGS)                                         \
		-t $(BUILD_IMAGE)                                             \
		-f Dockerfile.build .                                         \
		$(VERBOSE_OUTPUT)
	echo "$(BUILD_IMAGE)" > $@
	docker images -q $(BUILD_IMAGE) >> $@

# Rules for all bin/$(FAKEVER)/$(ARCH)/$(BINARY)
GO_BINARIES = $(addprefix bin/$(FAKEVER)/$(ARCH)/,$(BINARIES))
define GO_BINARIES_RULE
$(GO_BINARIES): build-dirs $(BUILD_IMAGE_BUILDSTAMP)
	@echo "building : $$@"
	docker run                                                               \
	    $(DOCKER_RUN_FLAGS)                                                  \
	    --sig-proxy=true                                                     \
	    -u $$$$(id -u):$$$$(id -g)                                           \
	    -v $$$$(pwd)/.go:/go                                                 \
	    -v $$$$(pwd):/go/src/$(PKG)                                          \
	    -v $$$$(pwd)/bin/$(FAKEVER)/$(ARCH):/go/bin                          \
	    -v $$$$(pwd)/bin/$(FAKEVER)/$(ARCH):/go/bin/linux_$(ARCH)            \
	    -v $$$$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static  \
	    -w /go/src/$(PKG)                                                    \
	    $(BUILD_IMAGE)                                                       \
	    /bin/sh -c "                                                         \
	        ARCH=$(ARCH)                                                     \
	        VERSION=$(VERSION_BASE)-$(FAKEVER)                               \
	        PKG=$(PKG)                                                       \
	        ./build/build.sh                                                 \
	    "
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
test: build-dirs $(BUILD_IMAGE_BUILDSTAMP)
	docker run                                                             \
	    $(DOCKER_RUN_FLAGS)                                                \
	    --sig-proxy=true                                                   \
	    -u $$(id -u):$$(id -g)                                             \
	    -v $$(pwd)/.go:/go                                                 \
	    -v $$(pwd):/go/src/$(PKG)                                          \
	    -v $$(pwd)/bin/$(ARCH):/go/bin                                     \
	    -v $$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static  \
	    -w /go/src/$(PKG)                                                  \
	    $(BUILD_IMAGE)                                                     \
	    /bin/sh -c "                                                       \
	        ./build/test.sh $(SRC_DIRS)                                    \
	    "

# Miscellaneous rules
.PHONY: version
version:
	@echo $(VERSION_BASE)

.PHONY: build-dirs
build-dirs:
	$(foreach FAKEVER,$(FAKE_VERSIONS),eval mkdir -p bin/$(FAKEVER)/$(ARCH);)
	mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(ARCH)

.PHONY: clean
clean: container-clean bin-clean

.PHONY: container-clean
container-clean:
	rm -f .*-container .*-dockerfile .*-push

.PHONY: bin-clean
bin-clean:
	rm -rf .go bin

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
