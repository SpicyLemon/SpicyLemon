#!/usr/bin/make -f
export GO111MODULE=on

GO ?= go
BUILDDIR ?= $(CURDIR)/build

ldflags = -w -s
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

build_flags = -mod=readonly -tags '$(build_tags)' -ldflags '$(ldflags)' -trimpath
build_flags += $(BUILD_FLAGS)
build_flags := $(strip $(build_flags))

go.sum:
	$(GO) mod verify
	$(GO) mod tidy

build: go.sum
	mkdir -p $(BUILDDIR)
	$(GO) build -o $(BUILDDIR)/ $(build_flags) .

install: go.sum
	$(GO) install $(build_flags) .

clean:
	rm -rf $(BUILDDIR)/*

.PHONY: go.sum build install clean

ifeq (,$(PACKAGES))
  PACKAGES := ./...
endif

test: go.sum
	$(GO) test -mod=readonly $(ARGS) -tags='$(TAGS)' $(PACKAGES)

.PHONY: test
