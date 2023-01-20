
GO ?= go
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

VERSION ?= $(shell git describe --dirty --always --tags | sed 's/-/./g')
GO_LDFLAGS := -ldflags '-s -w' -ldflags '-X github.com/alauda/kube-supv/version.BuildVersion=$(VERSION)'
PLATFORMS ?= linux_amd64 linux_arm64

.PHONY: all
all: fmt vet build

.PHONY: build
build: go.build.$(OS)_$(ARCH)

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: build.all
build.all: fmt vet $(foreach p,$(PLATFORMS),$(addprefix go.build., $(p)))

.PHONY: go.build.%
go.build.%:
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval OUTPUT_DIR := _output/$(OS)/$(ARCH))
	mkdir -p "$(OUTPUT_DIR)"
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -buildmode=pie $(GO_LDFLAGS) -o '$(OUTPUT_DIR)/kubesupv' .

.PHONY: package
package: build.all tar.all

.PHONY: tar.all
tar.all: $(foreach p,$(PLATFORMS),$(addprefix tar., $(p)))

.PHONY: tar.%
tar.%:
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval OUTPUT_DIR := _output/$(OS)/$(ARCH))
	tar czvf _output/kubesupv-$(OS)-$(ARCH).tgz -C $(OUTPUT_DIR) kubesupv

gen: controller-gen
	$(CONTROLLER_GEN) paths="./..." crd object output:crd:artifacts:config=config/crd

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
